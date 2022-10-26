package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"

	"github.com/evmos/evmos/v9/ibc"
	"github.com/evmos/evmos/v9/x/erc20/types"
	recoverytypes "github.com/evmos/evmos/v9/x/recovery/types"
)

// OnRecvPacket has been modified to use ConvertCoin.
// ConvertCoin converts IBC Coins to their ERC-20 Representations,
// given that the Token Pair is registered through governance.
// This conversion happens after the original IBC transfer.
func (k Keeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ack exported.Acknowledgement,
) exported.Acknowledgement {
	// Setup for OnRecvPacket:
	// - Get erc20 parameters for keeper
	// - Get sender/recipient addresses of transfer in `evmos1` and the original bech32 format
	// - Get denomination of packet's IBC transfer
	erc20Params := k.GetParams(ctx)
	sender, recipient, senderBech32, recipientBech32, err := ibc.GetTransferSenderRecipient(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}
	denom, err := ibc.GetTransferDenomination(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// Check and return error ACK if:
	// - Sender or recipient addresses are blocked
	if k.bankKeeper.BlockedAddr(sender) || k.bankKeeper.BlockedAddr(recipient) {
		return channeltypes.NewErrorAcknowledgement(
			sdkerrors.Wrapf(
				recoverytypes.ErrBlockedAddress,
				"sender (%s) or recipient (%s) address are in the deny list for sending and receiving transfers",
				senderBech32, recipientBech32,
			),
		)
	}

	// Return acknowledgement if:
	// - ERC20s are disabled
	// - The denomination is not registered
	if !(erc20Params.GetEnableErc20()) || !k.IsDenomRegistered(ctx, denom) {
		return ack
	}

	// Inward conversion, concerned only with IBC Coins:
	// (ERC20s would be registered with their contract address, would fail here)
	if k.IsDenomRegistered(ctx, denom) {
		stringAmount, err := ibc.GetTransferAmount(packet)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}
		intAmount, ok := sdk.NewIntFromString(stringAmount)
		if !ok {
			return channeltypes.NewErrorAcknowledgement(
				sdkerrors.Wrapf(
					errortypes.ErrInvalidCoins,
					"invalid amount: %s", stringAmount,
				),
			)
		}
		// Build coin for transfer
		coin := sdk.NewCoin(denom, intAmount)
		// Build MsgConvertCoin, from recipient to recipient since IBC transfer already occurred
		msg := types.NewMsgConvertCoin(coin, common.BytesToAddress(recipient.Bytes()), recipient)
		// Use MsgConvertCoin to convert the Cosmos Coin to an ERC20
		if _, err = k.ConvertCoin(sdk.WrapSDKContext(ctx), msg); err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}
		return ack
	}

	// Return acknowledgement:
	return ack
}
