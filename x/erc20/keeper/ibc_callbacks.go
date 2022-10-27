package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"

	"github.com/evmos/evmos/v9/ibc"
	"github.com/evmos/evmos/v9/x/erc20/types"
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
	// - Get denomination of packet's IBC transfer

	denom, err := ibc.GetTransferDenomination(packet)
	if err != nil {
		// NOTE: shouldn't occur, as the transfer has already been validated and
		// processed by the IBC transfer module
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// Return acknowledgement and continue with the next layer of the IBC middleware
	// stack if if:
	// - ERC20s are disabled
	// - The denomination is not registered as ERC20
	erc20Params := k.GetParams(ctx)
	if !erc20Params.EnableErc20 || !k.IsDenomRegistered(ctx, denom) {
		return ack
	}

	// Inward conversion, concerned only with IBC Coins:
	// (ERC20s would be registered with their contract address, would fail here)

	// FIXME: remove this as it's unmarshaling the IBC packet a second time
	stringAmount, err := ibc.GetTransferAmount(packet)
	if err != nil {
		// NOTE: shouldn't occur, as the transfer has already been validated and
		// processed by the IBC transfer module
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// FIXME: already done in GetTransferAmount, avoid dups
	intAmount, _ := sdk.NewIntFromString(stringAmount)

	// Build coin for transfer
	coin := sdk.NewCoin(denom, intAmount)

	// FIXME: remove this as it's unmarshaling the IBC packet a THIRD time!
	_, recipient, _, _, err := ibc.GetTransferSenderRecipient(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// Build MsgConvertCoin, from recipient to recipient since IBC transfer already occurred
	msg := types.NewMsgConvertCoin(coin, common.BytesToAddress(recipient.Bytes()), recipient)
	// Use MsgConvertCoin to convert the Cosmos Coin to an ERC20
	if _, err = k.ConvertCoin(sdk.WrapSDKContext(ctx), msg); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	return ack
}
