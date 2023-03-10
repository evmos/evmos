// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"

	"github.com/evmos/evmos/v12/ibc"
	"github.com/evmos/evmos/v12/x/claims/types"
)

// OnAcknowledgementPacket performs an IBC send callback. Once a user submits an
// IBC transfer to a recipient in the destination chain and the transfer
// acknowledgement package is received, the claimable amount for the senders's
// claims record `ActionIBCTransfer` is claimed and transferred to the sender
// address.
// The function performs a no-op if claims are disabled globally, acknowledgment
// failed, or if the sender has no claims record.
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
		return errorsmod.Wrapf(errortypes.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}

	// no-op if the acknowledgement is an error ACK
	if !ack.Success() {
		return nil
	}

	sender, _, _, _, err := ibc.GetTransferSenderRecipient(packet)
	if err != nil {
		return err
	}

	// Get claims record and return with no-op if sender doesn't have one
	claimsRecord, found := k.GetClaimsRecord(ctx, sender)
	if !found {
		return nil
	}

	// claim IBC transfer action
	_, err = k.ClaimCoinsForAction(ctx, sender, claimsRecord, types.ActionIBCTransfer, params)
	if err != nil {
		return err
	}

	return nil
}

// OnRecvPacket performs an IBC receive callback. Once a user receives an IBC
// transfer from a counterparty chain and the transfer is successful, the
// claimable amount for the receiver's claims record `ActionIBCTransfer` is
// claimed and transferred to the receivers address.
// Additionally, if the sender address is a Cosmos Hub or Osmosis address with
// an airdrop allocation, the claims record is merged with the recipient's
// claims record.
// The function performs a no-op if claims are disabled globally or if the
// sender has no claims record.
func (k Keeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ack exported.Acknowledgement,
) exported.Acknowledgement {
	logger := k.Logger(ctx)
	params := k.GetParams(ctx)

	// short (no-op) circuit by returning original ACK in case claims are not active
	if !params.IsClaimsActive(ctx.BlockTime()) {
		return ack
	}

	// Get bech32 address from the counterparty and change the bech32 human
	// readable prefix (HRP) of the sender to `evmos1`
	sender, recipient, senderBech32, recipientBech32, err := ibc.GetTransferSenderRecipient(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// return error ACK for blocked sender and recipient addresses
	if k.bankKeeper.BlockedAddr(sender) || k.bankKeeper.BlockedAddr(recipient) {
		return channeltypes.NewErrorAcknowledgement(
			errorsmod.Wrapf(
				errortypes.ErrUnauthorized,
				"sender (%s) or recipient (%s) address are in the deny list for sending and receiving transfers",
				senderBech32, recipientBech32,
			),
		)
	}

	senderClaimsRecord, senderRecordFound := k.GetClaimsRecord(ctx, sender)

	if senderRecordFound && senderClaimsRecord.HasClaimedAction(types.ActionIBCTransfer) {
		// short-circuit, perform no-op if the IBC action has already been completed
		return ack
	}

	sameAddress := sender.Equals(recipient)
	fromEVMChain := params.IsEVMChannel(packet.DestinationChannel)

	// If the packet is sent from a non-EVM chain, the sender address is not an
	// ethereum key (i.e. `ethsecp256k1`). Thus, if `sameAddress` is true, the
	// recipient address must be a non-ethereum key as well, which is not
	// supported on Evmos. To prevent funds getting stuck, return an error, unless
	// the destination channel from a connection to a chain is EVM-compatible or
	// supports ethereum keys (eg: Cronos, Injective).
	if sameAddress && !fromEVMChain {
		switch {
		case senderRecordFound && !senderClaimsRecord.HasClaimedAny():
			// secp256k1 key from sender/recipient has no claimed actions
			// -> return error acknowledgement to prevent funds from getting stuck
			return channeltypes.NewErrorAcknowledgement(
				errorsmod.Wrapf(
					types.ErrKeyTypeNotSupported, "receiver address %s is not a valid ethereum address", recipientBech32,
				),
			)
		default:
			// sender/recipient has funds stuck -> return ack to trigger withdrawal
			return ack
		}
	}

	// return original ACK in case the destination channel is not authorized
	if !params.IsAuthorizedChannel(packet.DestinationChannel) {
		return ack
	}

	recipientClaimsRecord, recipientRecordFound := k.GetClaimsRecord(ctx, recipient)

	amt, err := ibc.GetTransferAmount(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	isTriggerAmt := amt == types.IBCTriggerAmt

	switch {
	// Cases with SenderRecordFound.
	// They require a merge or migration of claims records. To prevent this
	// happening by accident, they are only executed, when the sender transfers
	// the specified IBCTriggerAmt.
	case senderRecordFound && recipientRecordFound && !sameAddress && isTriggerAmt:
		// case 1: both sender and recipient are distinct and have a claims record
		// -> merge sender's record with the recipient's record and claim actions that
		// have already been claimed by one or the other
		recipientClaimsRecord, err = k.MergeClaimsRecords(ctx, recipient, senderClaimsRecord, recipientClaimsRecord, params)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}

		// update the recipient's record with the new merged one and delete the
		// sender's record
		k.SetClaimsRecord(ctx, recipient, recipientClaimsRecord)
		k.DeleteClaimsRecord(ctx, sender)
		logger.Debug(
			"merged sender and receiver claims records",
			"sender", senderBech32,
			"receiver", recipientBech32,
			"total-claimable", senderClaimsRecord.InitialClaimableAmount.Add(recipientClaimsRecord.InitialClaimableAmount).String(),
		)
	case senderRecordFound && !recipientRecordFound && isTriggerAmt:
		// case 2: only the sender has a claims record
		// -> migrate the sender record to the recipient address and claim IBC action

		claimedAmt := sdk.ZeroInt() //nolint
		claimedAmt, err = k.ClaimCoinsForAction(ctx, recipient, senderClaimsRecord, types.ActionIBCTransfer, params)

		// if the transfer fails or the claimable amount is 0 (eg: action already
		// completed), don't perform a state migration
		if err != nil || claimedAmt.IsZero() {
			break
		}

		// delete the claims record from sender
		// NOTE: claim record is migrated to the recipient in ClaimCoinsForAction
		k.DeleteClaimsRecord(ctx, sender)

		logger.Debug(
			"migrated sender claims record to receiver",
			"sender", senderBech32,
			"receiver", recipientBech32,
			"total-claimable", senderClaimsRecord.InitialClaimableAmount.String(),
		)
	// Cases without SenderRecordFound
	case !senderRecordFound && recipientRecordFound,
		sameAddress && fromEVMChain && recipientRecordFound:
		// case 3: only the recipient has a claims record -> only claim IBC transfer action
		_, err = k.ClaimCoinsForAction(ctx, recipient, recipientClaimsRecord, types.ActionIBCTransfer, params)
	case !senderRecordFound && !recipientRecordFound:
		// case 4: neither the sender or recipient have a claims record
		// -> perform a no-op by returning the original success acknowledgement
		return ack
	}

	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// return the original success acknowledgement
	return ack
}
