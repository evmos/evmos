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

	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v14/ibc"
	"github.com/evmos/evmos/v14/x/erc20/types"
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

	if !k.IsERC20Enabled(ctx) {
		return ack
	}

	// Get addresses in `evmos1` and the original bech32 format
	sender, recipient, _, _, err := ibc.GetTransferSenderRecipient(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	claimsParams := k.claimsKeeper.GetParams(ctx)

	// if sender == recipient, and is not from an EVM Channel recovery was executed
	if sender.Equals(recipient) && !claimsParams.IsEVMChannel(packet.DestinationChannel) {
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
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	if coin.Denom == bondDenom {
		// no-op, received coin is the staking denomination
		return ack
	}

	pairID := k.GetTokenPairID(ctx, coin.Denom)
	if len(pairID) == 0 {
		// short-circuit: if the denom is not registered, conversion will fail
		// so we can continue with the rest of the stack
		return ack
	}

	pair, _ := k.GetTokenPair(ctx, pairID)
	if !pair.Enabled {
		// no-op: continue with the rest of the stack without conversion
		return ack
	}

	// Instead of converting just the received coins, convert the whole user balance
	// which includes the received coins.
	balance := k.bankKeeper.GetBalance(ctx, recipient, coin.Denom)

	// Build MsgConvertCoin, from recipient to recipient since IBC transfer already occurred
	msg := types.NewMsgConvertCoin(balance, common.BytesToAddress(recipient.Bytes()), recipient)

	// NOTE: we don't use ValidateBasic the msg since we've already validated
	// the ICS20 packet data

	// Use MsgConvertCoin to convert the Cosmos Coin to an ERC20
	if _, err = k.ConvertCoin(sdk.WrapSDKContext(ctx), msg); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
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
	sender, err := sdk.AccAddressFromBech32(data.Sender)
	if err != nil {
		return err
	}

	// use a zero gas config to avoid extra costs for the relayers
	ctx = ctx.
		WithKVGasConfig(storetypes.GasConfig{}).
		WithTransientKVGasConfig(storetypes.GasConfig{})

	// assume that all module accounts on Evmos need to have their tokens in the
	// IBC representation as opposed to ERC20
	senderAcc := k.accountKeeper.GetAccount(ctx, sender)
	if types.IsModuleAccount(senderAcc) {
		return nil
	}

	coin := ibc.GetSentCoin(data.Denom, data.Amount)

	// check if the coin is a native staking token
	bondDenom := k.stakingKeeper.BondDenom(ctx)
	if coin.Denom == bondDenom {
		// no-op, received coin is the staking denomination
		return nil
	}

	params := k.GetParams(ctx)
	if !params.EnableErc20 || !k.IsDenomRegistered(ctx, coin.Denom) {
		// no-op, ERC20s are disabled or the denom is not registered
		return nil
	}

	msg := types.NewMsgConvertCoin(coin, common.BytesToAddress(sender), sender)

	// NOTE: we don't use ValidateBasic the msg since we've already validated the
	// fields from the packet data

	// convert Coin to ERC20
	if _, err = k.ConvertCoin(sdk.WrapSDKContext(ctx), msg); err != nil {
		return err
	}

	defer func() {
		telemetry.IncrCounter(1, types.ModuleName, "ibc", "error", "total")
	}()

	return nil
}
