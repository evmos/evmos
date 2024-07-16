// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"strings"

	errorsmod "cosmossdk.io/errors"

	"github.com/armon/go-metrics"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/evmos/evmos/v18/ibc"
	"github.com/evmos/evmos/v18/x/erc20/types"
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
	// If ERC20 module is disabled no-op
	if !k.IsERC20Enabled(ctx) {
		return ack
	}

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

	sender, recipient, _, _, err := ibc.GetTransferSenderRecipient(data)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	evmParams := k.evmKeeper.GetParams(ctx)

	// If we received an IBC message from a non-EVM channel,
	// the sender and receiver accounts should be different.
	//
	// If its the same, users can have their funds stuck since they don't have access
	// to the same priv key. Return an error to prevent this from happening.
	//
	// This is an assumption to prevent possibly wrong transactions.
	if sender.Equals(recipient) && !evmParams.IsEVMChannel(packet.DestinationChannel) {
		return channeltypes.NewErrorAcknowledgement(types.ErrInvalidIBC)
	}

	receiverAcc := k.accountKeeper.GetAccount(ctx, recipient)

	// return acknowledgement without conversion if receiver is a module account
	if types.IsModuleAccount(receiverAcc) {
		return ack
	}

	// parse the transferred denom
	coin := ibc.GetReceivedCoin(
		packet.SourcePort, packet.SourceChannel,
		packet.DestinationPort, packet.DestinationChannel,
		data.Denom, data.Amount,
	)

	// If the coin denom starts with `factory/` then it is a token factory coin, and we should not convert it
	// NOTE: Check https://docs.osmosis.zone/osmosis-core/modules/tokenfactory/ for more information
	if strings.HasPrefix(data.Denom, "factory/") {
		return ack
	}

	// check if the coin is a native staking token
	if coin.Denom == evmParams.EvmDenom {
		// no-op, received coin is the staking denomination
		return ack
	}

	pairID := k.GetTokenPairID(ctx, coin.Denom)
	pair, found := k.GetTokenPair(ctx, pairID)
	switch {
	// Case 1. token pair is not registered and is a single hop IBC Coin
	// by checking the prefix we ensure that only coins not native from this chain are evaluated.
	// IsNativeFromSourceChain will check if the coin is native from the source chain.
	// If the coin denom starts with `factory/` then it is a token factory coin, and we should not convert it
	// NOTE: Check https://docs.osmosis.zone/osmosis-core/modules/tokenfactory/ for more information
	case !found && strings.HasPrefix(coin.Denom, "ibc/") && ibc.IsBaseDenomFromSourceChain(data.Denom):
		tokenPair, err := k.RegisterERC20Extension(ctx, coin.Denom)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}

		ctx.EventManager().EmitEvents(
			sdk.Events{
				sdk.NewEvent(
					types.EventTypeRegisterERC20Extension,
					sdk.NewAttribute(types.AttributeCoinSourceChannel, packet.SourceChannel),
					sdk.NewAttribute(types.AttributeKeyERC20Token, tokenPair.Erc20Address),
					sdk.NewAttribute(types.AttributeKeyCosmosCoin, tokenPair.Denom),
				),
			},
		)
		return ack

	// Case 2. native ERC20 token
	case found && pair.IsNativeERC20():
		// Token pair is disabled -> return
		if !pair.Enabled {
			return ack
		}

		balance := k.bankKeeper.GetBalance(ctx, recipient, coin.Denom)
		if err := k.ConvertCoinNativeERC20(ctx, pair, balance.Amount, common.BytesToAddress(recipient.Bytes()), recipient); err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}

		// For now the only case we are interested in adding telemetry is a successful conversion.
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, "ibc", "on_recv", "total"},
			1,
			[]metrics.Label{
				telemetry.NewLabel("denom", coin.Denom),
				telemetry.NewLabel("source_channel", packet.SourceChannel),
				telemetry.NewLabel("source_port", packet.SourcePort),
			},
		)
	}

	return ack
}

// OnAcknowledgementPacket responds to the success or failure of a packet
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
// This function is only executed when IBC timeout or an Error ACK happens.
func (k Keeper) ConvertCoinToERC20FromPacket(ctx sdk.Context, data transfertypes.FungibleTokenPacketData) error {
	sender, err := sdk.AccAddressFromBech32(data.Sender)
	if err != nil {
		return err
	}

	pairID := k.GetTokenPairID(ctx, data.Denom)
	pair, found := k.GetTokenPair(ctx, pairID)
	if !found {
		// no-op, token pair is not registered
		return nil
	}

	coin := ibc.GetSentCoin(data.Denom, data.Amount)

	switch {

	// Case 1. if pair is native coin -> no-op
	case pair.IsNativeCoin():
		// no-op, received coin is a  native coin
		return nil

	// Case 2. if pair is native ERC20 -> unescrow
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

		// assume that all module accounts on Evmos need to have their tokens in the
		// IBC representation as opposed to ERC20
		senderAcc := k.accountKeeper.GetAccount(ctx, sender)
		if types.IsModuleAccount(senderAcc) {
			return nil
		}

		// Convert from Coin to ERC20
		if err := k.ConvertCoinNativeERC20(ctx, pair, coin.Amount, common.BytesToAddress(sender), sender); err != nil {
			// We want to record only the failed attempt to reconvert the coins during IBC.
			defer func() {
				telemetry.IncrCounter(1, types.ModuleName, "ibc", "error", "total")
			}()

			return err
		}
	}

	return nil
}
