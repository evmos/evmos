// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/armon/go-metrics"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/evmos/evmos/v16/utils"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v16/ibc"
	"github.com/evmos/evmos/v16/x/erc20/types"
)

// OnRecvPacket performs the ICS20 middleware receive callback for automatically
// converting an IBC Coin to their ERC20 representation.
// For the conversion to succeed, the IBC denomination must have previously been
// registered via governance. Note that the native staking denomination (e.g. "aevmos"),
// is excluded from the conversion.
//
// CONTRACT: This middleware MUST be executed transfer after the ICS20 OnRecvPacket
// Return acknowledgement and continue with the next layer of the IBC middleware
// stack if:
// - ERC20s are disabled
// - Denomination is native staking token
// - The base denomination is not registered as ERC20
func (k Keeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ack exported.Acknowledgement,
) exported.Acknowledgement {
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		// NOTE: shouldn't happen as the packet has already
		// been decoded on ICS20 transfer logic
		err = errorsmod.Wrapf(errortypes.ErrInvalidType, "cannot unmarshal ICS-20 transfer packet data")
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// use a zero gas config to avoid extra costs for the relayers
	ctx = ctx.
		WithKVGasConfig(storetypes.GasConfig{}).
		WithTransientKVGasConfig(storetypes.GasConfig{})

	// Get addresses in `evmos1` and the original bech32 format
	sender, recipient, _, _, err := ibc.GetTransferSenderRecipient(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	evmParams := k.evmKeeper.GetParams(ctx)

	// if sender == recipient, and is not from an EVM Channel recovery was executed
	if sender.Equals(recipient) && !evmParams.IsEVMChannel(packet.DestinationChannel) {
		// Continue to the next IBC middleware by returning the original ACK.
		return ack
	}

	senderAcc := k.accountKeeper.GetAccount(ctx, sender)

	// return acknoledgement without conversion if sender is a module account
	if types.IsModuleAccount(senderAcc) {
		return ack
	}

	// parse the transferred denom
	coin := ibc.GetReceivedCoin(
		packet.SourcePort, packet.SourceChannel,
		packet.DestinationPort, packet.DestinationChannel,
		data.Denom, data.Amount,
	)

	// check if the coin is a native staking token
	if coin.Denom == evmParams.EvmDenom {
		// no-op, received coin is the staking denomination
		return ack
	}

	pairID := k.GetTokenPairID(ctx, coin.Denom)
	pair, found := k.GetTokenPair(ctx, pairID)
	switch {
	// Case 1. token pair is not registered and is a single hop IBC Coin
	case !found && ibc.IsSingleHop(coin.Denom):
		contractAddr, err := utils.GetIBCDenomAddress(coin.Denom)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}

		found := evmParams.IsPrecompileRegistered(contractAddr.String())
		if found {
			return ack
		}

		if err := k.RegisterPrecompileForCoin(ctx, coin.Denom, contractAddr); err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}
		return ack

	// Case 2. native ERC20 token
	case pair.IsNativeERC20():
		// ERC20 module or token pair is disabled -> return
		if !k.IsERC20Enabled(ctx) || !pair.Enabled {
			return ack
		}

		msgConvert := types.NewMsgConvertERC20(coin.Amount, recipient, pair.GetERC20Contract(), common.BytesToAddress(sender))
		// Convert from Coin to ERC20
		_, err := k.ConvertERC20(ctx, msgConvert)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}

	// TODO: Is the default just an ack or an error ?
	default:
		return ack
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, "ibc", "on_recv", "total"},
			1,
			[]metrics.Label{
				telemetry.NewLabel("denom", coin.Denom),
				telemetry.NewLabel("source_channel", packet.SourceChannel),
				telemetry.NewLabel("source_port", packet.SourcePort),
			},
		)
	}()

	return ack
}

// OnAcknowledgementPacket responds to the the success or failure of a packet
// acknowledgement written on the receiving chain. If the acknowledgement was a
// success then nothing occurs. If the acknowledgement failed, then the sender
// is refunded and then the IBC Coins are converted to ERC20.
func (k Keeper) OnAcknowledgementPacket(
	ctx sdk.Context, _ channeltypes.Packet,
	data transfertypes.FungibleTokenPacketData,
	ack channeltypes.Acknowledgement,
) error {
	switch ack.Response.(type) {
	case *channeltypes.Acknowledgement_Error:
		// convert the token from Cosmos Coin to its ERC20 representation
		return k.ConvertCoinToERC20FromPacket(ctx, data)
	default:
		// the acknowledgement succeeded on the receiving chain so nothing needs to
		// be executed and no error needs to be returned
		return nil
	}
}

// OnTimeoutPacket converts the IBC coin to ERC20 after refunding the sender
// since the original packet sent was never received and has been timed out.
func (k Keeper) OnTimeoutPacket(ctx sdk.Context, _ channeltypes.Packet, data transfertypes.FungibleTokenPacketData) error {
	return k.ConvertCoinToERC20FromPacket(ctx, data)
}

// ConvertCoinToERC20FromPacket converts the IBC coin to ERC20 after refunding the sender
func (k Keeper) ConvertCoinToERC20FromPacket(ctx sdk.Context, data transfertypes.FungibleTokenPacketData) error {
	pairID := k.GetTokenPairID(ctx, data.Denom)
	pair, found := k.GetTokenPair(ctx, pairID)
	if !found {
		// no-op, token pair is not registered
		return nil
	}

	coin := ibc.GetSentCoin(data.Denom, data.Amount)
	// check if the coin is a native staking token
	bondDenom := k.stakingKeeper.BondDenom(ctx)

	switch {
	// Case 1. if pair is native denomination -> no-op
	case coin.Denom == bondDenom:
		// no-op, received coin is the staking denomination
		return nil
	// Case 2. if pair is native coin -> no-op
	case pair.IsNativeCoin():
		// no-op, received coin is the native denomination
		return nil

	// Case 3. if pair is native ERC20 -> unescrow
	case pair.IsNativeERC20():
		// use a zero gas config to avoid extra costs for the relayers
		ctx = ctx.
			WithKVGasConfig(storetypes.GasConfig{}).
			WithTransientKVGasConfig(storetypes.GasConfig{})

		params := k.GetParams(ctx)
		if !params.EnableErc20 || !k.IsDenomRegistered(ctx, coin.Denom) {
			// no-op, ERC20s are disabled or the denom is not registered
			return nil
		}

		receiver, err := sdk.AccAddressFromBech32(data.Receiver)
		if err != nil {
			return err
		}

		sender, err := sdk.AccAddressFromBech32(data.Sender)
		if err != nil {
			return err
		}

		// assume that all module accounts on Evmos need to have their tokens in the
		// IBC representation as opposed to ERC20
		senderAcc := k.accountKeeper.GetAccount(ctx, sender)
		if types.IsModuleAccount(senderAcc) {
			return nil
		}

		if err := k.convertCoinNativeERC20(ctx, pair, coin, common.BytesToAddress(receiver), sender); err != nil {
			return err
		}
	}

	defer func() {
		telemetry.IncrCounter(1, types.ModuleName, "ibc", "error", "total")
	}()

	return nil
}
