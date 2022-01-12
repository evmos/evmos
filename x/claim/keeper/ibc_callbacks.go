package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tharsis/evmos/x/claim/types"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
)

// OnRecvPacket performs an IBC callback.
func (k Keeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ack exported.Acknowledgement,
) exported.Acknowledgement {
	if !ack.Success() || !k.IsTransferHooksEnabled(ctx) {
		return ack
	}

	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
		return channeltypes.NewErrorAcknowledgement(err.Error())
	}

	sender, err := sdk.AccAddressFromBech32(data.Sender)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err.Error())
	}

	recipient, err := sdk.AccAddressFromBech32(data.Receiver)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err.Error())
	}

	// migrate the claim record to recipient address
	claimRecord, found := k.GetClaimRecord(ctx, sender)
	if found {
		k.SetClaimRecord(ctx, recipient, claimRecord)
		k.DeleteClaimRecord(ctx, sender)
	}

	// claim IBC action
	if _, err := k.ClaimCoinsForAction(ctx, recipient, claimRecord, types.ActionIBCTransfer); err != nil {
		return channeltypes.NewErrorAcknowledgement(err.Error())
	}

	// return the original success acknowledgement
	return ack
}

func (k Keeper) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
) error {
	var ack channeltypes.Acknowledgement
	if err := transfertypes.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}

	if !ack.Success() || !k.IsTransferHooksEnabled(ctx) {
		return nil
	}

	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	sender, err := sdk.AccAddressFromBech32(data.Sender)
	if err != nil {
		return err
	}

	claimRecord, found := k.GetClaimRecord(ctx, sender)
	if !found {
		return nil
	}

	// claim IBC action
	if _, err := k.ClaimCoinsForAction(ctx, sender, claimRecord, types.ActionIBCTransfer); err != nil {
		return err
	}

	return nil
}
