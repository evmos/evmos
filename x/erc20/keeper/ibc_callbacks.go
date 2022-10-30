package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ethereum/go-ethereum/common"

	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"

	evmos "github.com/evmos/evmos/v9/types"
	"github.com/evmos/evmos/v9/x/erc20/types"
)

// OnRecvPacket performs an IBC receive callback. Once a user receives
// an IBC transfer and the transfer is successful, the IBC Coins gets converted
// to their ERC-20 Representations, given that the Token Pair is
// registered through governance.
func (k Keeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ack exported.Acknowledgement,
) exported.Acknowledgement {
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		err = sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "cannot unmarshal ICS-20 transfer packet data")
		return channeltypes.NewErrorAcknowledgement(err)
	}

	denomTrace := transfertypes.ParseDenomTrace(data.Denom)
	denom := denomTrace.GetBaseDenom()

	// Return acknowledgement and continue with the next layer of the IBC middleware
	// stack if if:
	// - ERC20s are disabled
	// - The base denomination is not registered as ERC20
	erc20Params := k.GetParams(ctx)
	if !erc20Params.EnableErc20 {
		return ack
	}

	// TODO: check if the token denom is source?
	if !k.IsDenomRegistered(ctx, denom) {
		return ack
	}

	amount, _ := sdk.NewIntFromString(data.Amount)

	// Setup for OnRecvPacket:
	// - Get erc20 parameters for keeper
	// - Get denomination of packet's IBC transfer

	// TODO:
	// denom, err := ibc.GetTransferDenomination(packet)
	// if err != nil {
	// 	// NOTE: shouldn't occur, as the transfer has already been validated and
	// 	// processed by the IBC transfer module
	// 	return channeltypes.NewErrorAcknowledgement(err)
	// }

	// Inward conversion, concerned only with IBC Coins:
	// (ERC20s would be registered with their contract address, would fail here)

	recipient, err := evmos.GetEvmosAddressFromBech32(data.Receiver)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(
			sdkerrors.Wrap(err, "invalid recipient"),
		)
	}

	// Build coin for transfer without validating it again
	coin := sdk.Coin{Denom: denom, Amount: amount}

	// Build MsgConvertCoin, from recipient to recipient since IBC transfer already occurred
	msg := types.NewMsgConvertCoin(coin, common.BytesToAddress(recipient.Bytes()), recipient)

	// NOTE: we don't use ValidateBasic the msg since we've already validated
	// the ICS20 packet data

	// Use MsgConvertCoin to convert the Cosmos Coin to an ERC20
	if _, err = k.ConvertCoin(sdk.WrapSDKContext(ctx), msg); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	return ack
}
