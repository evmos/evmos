package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/tharsis/evmos/v2/ibc"
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
	logger := k.Logger(ctx)
	params := k.GetParams(ctx)

	// short (no-op) circuit by returning original ACK in case the claim is not active
	if !params.IsClaimsActive(ctx.BlockTime()) {
		return ack
	}

	sender, recipient, senderBech32, recipientBech32, err := ibc.GetTransferSenderRecipient(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err.Error())
	}

	// return error ACK for blocked sender and recipient addresses
	if k.bankKeeper.BlockedAddr(sender) || k.bankKeeper.BlockedAddr(recipient) {
		return channeltypes.NewErrorAcknowledgement(
			sdkerrors.Wrapf(
				sdkerrors.ErrUnauthorized,
				"sender (%s) or recipient (%s) address are in the deny list for sending and receiving transfers",
				senderBech32, recipientBech32,
			).Error(),
		)
	}

	senderClaimsRecord, senderRecordFound := k.GetClaimsRecord(ctx, sender)

	sameAddress := sender.Equals(recipient)
	fromEVMChain := params.IsEVMChannel(packet.DestinationChannel)

	// NOTE: we know that the connected chains from the authorized IBC channels
	// don't support ethereum keys (i.e `ethsecp256k1`). Thus, so we return an error,
	// unless the destination channel from a connection to a chain that is EVM-compatible
	// or supports ethereum keys (eg: Cronos, Injective).
	if sameAddress && !fromEVMChain {
		switch {
		// case 1: secp256k1 key from sender/recipient has no claimed actions -> error ACK to prevent funds from getting stuck
		case senderRecordFound && !senderClaimsRecord.HasClaimedAny():
			return channeltypes.NewErrorAcknowledgement(
				sdkerrors.Wrapf(
					evmos.ErrKeyTypeNotSupported, "receiver address %s is not a valid ethereum address", recipientBech32,
				).Error(),
			)
		default:
			// case 2: sender/recipient has funds stuck -> error acknowledgement to prevent more transferred tokens from
			// getting stuck while we implement IBC withdrawals
			return channeltypes.NewErrorAcknowledgement(
				sdkerrors.Wrapf(
					evmos.ErrKeyTypeNotSupported,
					"reverted transfer to unsupported address %s to prevent more funds from getting stuck",
					recipientBech32,
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
	case senderRecordFound && recipientRecordFound && !sameAddress:
		// 1. Both sender and recipient (distinct addresses) have a claims record
		// Merge sender's record with the recipient's record and
		// claim actions that have been completed by one or the other
		recipientClaimsRecord, err = k.MergeClaimsRecords(ctx, recipient, senderClaimsRecord, recipientClaimsRecord, params)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err.Error())
		}

		// update the recipient's record with the new merged one, while deleting the
		// sender's record
		k.SetClaimsRecord(ctx, recipient, recipientClaimsRecord)
		// only delete if sender != recipient
		k.DeleteClaimsRecord(ctx, sender)
		logger.Debug(
			"merged sender and receiver claims records",
			"sender", senderBech32,
			"receiver", recipientBech32,
			"total-claimable", senderClaimsRecord.InitialClaimableAmount.Add(recipientClaimsRecord.InitialClaimableAmount).String(),
		)
	case senderRecordFound && !recipientRecordFound:
		// 2. Only the sender has a claims record.
		// Migrate the sender record to the recipient address
		k.SetClaimsRecord(ctx, recipient, senderClaimsRecord)
		k.DeleteClaimsRecord(ctx, sender)

		logger.Debug(
			"migrated sender claims record to receiver",
			"sender", senderBech32,
			"receiver", recipientBech32,
			"total-claimable", senderClaimsRecord.InitialClaimableAmount.String(),
		)

		// claim IBC action
		_, err = k.ClaimCoinsForAction(ctx, recipient, senderClaimsRecord, types.ActionIBCTransfer, params)
	case !senderRecordFound && recipientRecordFound,
		sameAddress && fromEVMChain && recipientRecordFound:
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

	sender, _, _, _, err := ibc.GetTransferSenderRecipient(packet)
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
