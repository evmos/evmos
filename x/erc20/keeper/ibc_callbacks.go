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
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/ibc"
	"github.com/evmos/evmos/v16/utils"
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
	sender, receiver, _, _, err := ibc.GetTransferSenderRecipient(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	senderAcc := k.accountKeeper.GetAccount(ctx, sender)
	// return acknowledgement without conversion if sender is a module account
	if types.IsModuleAccount(senderAcc) {
		return ack
	}

	// parse the transferred denom
	coin := ibc.GetReceivedCoin(
		packet.SourcePort, packet.SourceChannel,
		packet.DestinationPort, packet.DestinationChannel,
		data.Denom, data.Amount,
	)

	// If native coin just return sooner
	if coin.Denom == k.stakingKeeper.BondDenom(ctx) {
		return ack
	}

	pairID := k.GetTokenPairID(ctx, coin.Denom)
	pair, found := k.GetTokenPair(ctx, pairID)

	// TODO: Consider how it integrates with PFM.
	// Case 1 - token pair is not registered
	// Case 1.1 - coin is a native chain voucher and the token pair is not registered
	if !found && ibc.IsSingleHop(coin.Denom) {
		denomAddr, err := utils.GetIBCDenomAddress(coin.Denom)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}

		// TODO: Where are we populating the name, symbol and decimals for the metadata ?
		// For an IBC voucher will there be metadata already available in the bank ?
		// coinMetadata, found := k.bankKeeper.GetDenomMetaData(ctx, coin.Denom)
		coinMetadata, err := k.CreateERC20Metadata(ctx, coin.Denom, "", "", 0)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}

		// Truncate the base denom to 20 characters
		truncatedAddr := denomAddr[:20]
		coinMetadata.Base = common.BytesToAddress(truncatedAddr).String()

		tokenPair, err := k.RegisterTokenPairForNativeCoin(ctx, *coinMetadata)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}

		if err := k.RegisterPrecompileForCoin(ctx, *tokenPair); err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}

		return ack
	}

	// Case 2 - native ERC20 token
	if pair.IsNativeERC20() {
		// ERC20 module or token pair is disabled -> return
		if !k.IsERC20Enabled(ctx) || !pair.Enabled {
			return ack
		}

		msgConvert := types.NewMsgConvertERC20(coin.Amount, receiver, pair.GetERC20Contract(), common.BytesToAddress(sender))
		// Convert from Coin to ERC20
		_, err := k.ConvertERC20(ctx, msgConvert)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}
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
// success then nothing occurs.
func (k Keeper) OnAcknowledgementPacket(
	_ sdk.Context, _ channeltypes.Packet,
	_ transfertypes.FungibleTokenPacketData,
	ack channeltypes.Acknowledgement,
) error {
	switch ack.Response.(type) {
	case *channeltypes.Acknowledgement_Error:
		// TODO: We dont need to do anything here because there is no minting and burning happening ?
		return nil
	default:
		// the acknowledgement succeeded on the receiving chain so nothing needs to
		// be executed and no error needs to be returned
		return nil
	}
}

// OnTimeoutPacket converts the IBC coin to ERC20 after refunding the sender
// since the original packet sent was never received and has been timed out.
func (k Keeper) OnTimeoutPacket(_ sdk.Context, _ channeltypes.Packet, _ transfertypes.FungibleTokenPacketData) error {
	// TODO: We do nothing here because there is no burning / minting mechanism ?
	return nil
}
