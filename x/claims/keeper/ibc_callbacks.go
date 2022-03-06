package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	evmos "github.com/tharsis/evmos/v2/types"
	"github.com/tharsis/evmos/v2/x/claims/types"
)

// OnRecvPacket performs an IBC receive callback. It performs a no-op if
// claims are inactive
func (k Keeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ack exported.Acknowledgement,
) exported.Acknowledgement {
	params := k.GetParams(ctx)

	// short (no-op) circuit by returning original ACK in case the claim is not active
	if !params.IsClaimsActive(ctx.BlockTime()) {
		return ack
	}

	// unmarshal packet data to obtain the sender and recipient
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data")
		return channeltypes.NewErrorAcknowledgement(err.Error())
	}

	// validate the sender bech32 address from the counterparty chain
	bech32Prefix := strings.Split(data.Sender, "1")[0]
	if bech32Prefix == data.Sender {
		return channeltypes.NewErrorAcknowledgement(
			sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender: %s", data.Sender).Error(),
		)
	}

	senderBz, err := sdk.GetFromBech32(data.Sender, bech32Prefix)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(
			sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender %s, %s", data.Sender, err.Error()).Error(),
		)
	}

	// change the bech32 human readable prefix (HRP) of the sender to `evmos1`
	sender := sdk.AccAddress(senderBz)

	// obtain the evmos recipient address
	recipient, err := sdk.AccAddressFromBech32(data.Receiver)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(
			sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid receiver address %s", err.Error()).Error(),
		)
	}

	senderClaimsRecord, senderRecordFound := k.GetClaimsRecord(ctx, sender)

	// NOTE: we know that the connected chains from the authorized IBC channels
	// don't support ethereum keys (i.e `ethsecp256k1`). Thus, so we return an error,
	// unless the destination channel from a connection to a chain that is EVM-compatible
	// or supports ethereum keys (eg: Cronos, Injective).
	if sender.Equals(recipient) && !params.IsEVMChannel(packet.DestinationChannel) {
		switch {
		// case 1: secp256k1 key from sender/recipient has no claimed actions -> error ACK to prevent funds from getting stuck
		case senderRecordFound && !senderClaimsRecord.HasClaimedAny():
			return channeltypes.NewErrorAcknowledgement(
				sdkerrors.Wrapf(
					evmos.ErrKeyTypeNotSupported, "receiver address %s is not a valid ethereum address", data.Receiver,
				).Error(),
			)
		default:
			// case 2: sender/recipient has funds stuck -> error acknowledgement to prevent more transferred tokens from
			// getting stuck while we implement IBC withdrawals
			return channeltypes.NewErrorAcknowledgement(
				sdkerrors.Wrapf(
					evmos.ErrKeyTypeNotSupported,
					"reverted transfer to unsupported address %s to prevent more funds from getting stuck",
					data.Receiver,
				).Error(),
			)
		}
	}

	// return original ACK in case the destination channel is not authorized
	if !params.IsAuthorizedChannel(packet.DestinationChannel) {
		return ack
	}

	recipientClaimsRecord, recipientRecordFound := k.GetClaimsRecord(ctx, recipient)

	// handle the 4 cases for the recipient and sender claim records

	switch {
	case senderRecordFound && recipientRecordFound:
		// 1. Both sender and recipient have a claims record
		// Merge sender's record with the recipient's record and
		// claim actions that have been completed by one or the other
		recipientClaimsRecord, err = k.MergeClaimsRecords(ctx, recipient, senderClaimsRecord, recipientClaimsRecord, params)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err.Error())
		}

		// update the recipient's record with the new merged one, while deleting the
		// sender's record
		k.SetClaimsRecord(ctx, recipient, recipientClaimsRecord)
		k.DeleteClaimsRecord(ctx, sender)
	case senderRecordFound && !recipientRecordFound:
		// 2. Only the sender has a claims record.
		// Migrate the sender record to the recipient address
		k.SetClaimsRecord(ctx, recipient, senderClaimsRecord)
		k.DeleteClaimsRecord(ctx, sender)

		// claim IBC action
		_, err = k.ClaimCoinsForAction(ctx, recipient, senderClaimsRecord, types.ActionIBCTransfer, params)
	case !senderRecordFound && recipientRecordFound:
		// 3. Only the recipient has a claims record.
		// Only claim IBC transfer action
		_, err = k.ClaimCoinsForAction(ctx, recipient, recipientClaimsRecord, types.ActionIBCTransfer, params)
	case !senderRecordFound && !recipientRecordFound:
		// 4. Neither the sender or recipient have a claims record.
		// Perform a no-op by returning the  original success acknowledgement
		return ack
	}

	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err.Error())
	}

	// return the original success acknowledgement
	return ack
}

// OnAcknowledgementPacket claims the amount from the `ActionIBCTransfer` for
// the sender of the IBC transfer.
// The function performs a no-op if claims are disabled globally,
// acknowledgment failed, or if sender the sender has no claims record.
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
		// no-op. The user doesn't have a claim record so we don't need to perform
		// any claim
		return nil
	}

	// claim IBC transfer action
	_, err = k.ClaimCoinsForAction(ctx, sender, claimRecord, types.ActionIBCTransfer, params)
	if err != nil {
		return err
	}

	return nil
}
