package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/tharsis/evmos/x/claims/types"
)

// OnRecvPacket performs an IBC callback.
func (k Keeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ack exported.Acknowledgement,
) exported.Acknowledgement {
	params := k.GetParams(ctx)

	// short circuit in case claim is not active (no-op) or if the
	// acknowledgement is an error ACK
	if !ack.Success() || !params.IsClaimsActive(ctx.BlockTime()) {
		return ack
	}

	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data")
		return channeltypes.NewErrorAcknowledgement(err.Error())
	}

	// TODO: verify since sender will be from other chain?
	sender, err := sdk.AccAddressFromBech32(data.Sender)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err.Error())
	}

	recipient, err := sdk.AccAddressFromBech32(data.Receiver)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err.Error())
	}

	senderClaimRecord, senderRecordFound := k.GetClaimsRecord(ctx, sender)
	recipientClaimRecord, recipientRecordFound := k.GetClaimsRecord(ctx, recipient)

	switch {
	case senderRecordFound && recipientRecordFound:
		// claim already claimed actions (recipient) for sender
		// MERGE sender's record with the recipient's record

		// 2.1.1. if an action been claimed by recipient
		//   -> TODO: no-op? calculation gets messy since the airdrop is divided from the actions not claimed
		// 		-> TODO: we could also claim all the actions already claimed by the recipient
		// 2.1.2  if no action has been claimed -> add to total
		// if the recipient already has a claim record,
		// add the initial balance to the
	case senderRecordFound && !recipientRecordFound:
		// migrate sender record to recipient
		k.SetClaimsRecord(ctx, recipient, senderClaimRecord)
		k.DeleteClaimsRecord(ctx, sender)

		// claim IBC action
		_, err = k.ClaimCoinsForAction(ctx, recipient, senderClaimRecord, types.ActionIBCTransfer, params)
	case !senderRecordFound && recipientRecordFound:
		// claim IBC transfer action
		_, err = k.ClaimCoinsForAction(ctx, recipient, recipientClaimRecord, types.ActionIBCTransfer, params)
	case !senderRecordFound && !recipientRecordFound:
		// return original success acknowledgement
		return ack
	}

	if err != nil {
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
	params := k.GetParams(ctx)

	// short circuit in case claim is not active (no-op)
	if !params.IsClaimsActive(ctx.BlockTime()) {
		return nil
	}

	var ack channeltypes.Acknowledgement
	if err := transfertypes.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}

	// no-op if the acknowledgement is an error ACK
	if !ack.Success() {
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

	claimRecord, found := k.GetClaimsRecord(ctx, sender)
	if !found {
		// user doesn't have a claim record so we don't need to perform any claim
		return nil
	}

	// claim IBC action
	_, err = k.ClaimCoinsForAction(ctx, sender, claimRecord, types.ActionIBCTransfer, params)
	if err != nil {
		return err
	}

	return nil
}
