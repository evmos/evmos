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
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"

	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	evmtypes "github.com/evmos/evmos/v12/x/evm/types"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/evmos/evmos/v12/x/claims/types"
)

var (
	_ porttypes.ICS4Wrapper     = Keeper{}
	_ evmtypes.EvmHooks         = Hooks{}
	_ govtypes.GovHooks         = Hooks{}
	_ stakingtypes.StakingHooks = Hooks{}
)

// Hooks wrapper struct for the claim keeper
type Hooks struct {
	k Keeper
}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// AfterProposalVote is a wrapper for calling the Gov AfterProposalVote hook on
// the module keeper
func (h Hooks) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	h.k.AfterProposalVote(ctx, proposalID, voterAddr)
}

// AfterProposalVote is called after a vote on a proposal is cast. Once the vote
// is successfully included, the claimable amount for the user's claims record
// vote action is claimed and the transferred to the user address.
func (k Keeper) AfterProposalVote(ctx sdk.Context, _ uint64, voterAddr sdk.AccAddress) {
	params := k.GetParams(ctx)

	claimsRecord, found := k.GetClaimsRecord(ctx, voterAddr)
	if !found {
		return
	}

	_, err := k.ClaimCoinsForAction(ctx, voterAddr, claimsRecord, types.ActionVote, params)
	if err != nil {
		k.Logger(ctx).Error(
			"failed to claim Vote action",
			"address", voterAddr.String(),
			"error", err.Error(),
		)
	}
}

// AfterDelegationModified is a wrapper for calling the Staking AfterDelegationModified
// hook on the module keeper
func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	err := h.k.AfterDelegationModified(ctx, delAddr, valAddr)
	return err
}

// AfterDelegationModified is called after a delegation is modified. Once a user
// delegates their EVMOS tokens to a validator, the claimable amount for the
// user's claims record delegation action is claimed and transferred to the user
// address.
func (k Keeper) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, _ sdk.ValAddress) error {
	params := k.GetParams(ctx)

	claimsRecord, found := k.GetClaimsRecord(ctx, delAddr)
	if !found {
		return nil
	}

	_, err := k.ClaimCoinsForAction(ctx, delAddr, claimsRecord, types.ActionDelegate, params)
	if err != nil {
		k.Logger(ctx).Error(
			"failed to claim Delegation action",
			"address", delAddr.String(),
			"error", err.Error(),
		)
		return nil
	}
	return nil
}

// PostTxProcessing is a wrapper for calling the EVM PostTxProcessing hook on
// the module keeper
func (h Hooks) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	return h.k.PostTxProcessing(ctx, msg, receipt)
}

// PostTxProcessing implements the ethermint evm PostTxProcessing hook.
// After a EVM state transition is successfully processed, the claimable amount
// for the users's claims record evm action is claimed and transferred to the
// user address.
func (k Keeper) PostTxProcessing(ctx sdk.Context, msg core.Message, _ *ethtypes.Receipt) error {
	params := k.GetParams(ctx)
	fromAddr := sdk.AccAddress(msg.From().Bytes())

	claimsRecord, found := k.GetClaimsRecord(ctx, fromAddr)
	if !found {
		return nil
	}

	_, err := k.ClaimCoinsForAction(ctx, fromAddr, claimsRecord, types.ActionEVM, params)
	if err != nil {
		k.Logger(ctx).Error(
			"failed to claim EVM action",
			"address", fromAddr.String(),
			"error", err.Error(),
		)
	}

	return nil
}

// ________________________________________________________________________________________

// governance hooks
func (h Hooks) AfterProposalFailedMinDeposit(_ sdk.Context, _ uint64) {
}

func (h Hooks) AfterProposalVotingPeriodEnded(_ sdk.Context, _ uint64) {
}

func (h Hooks) AfterProposalSubmission(_ sdk.Context, _ uint64) {}

func (h Hooks) AfterProposalDeposit(_ sdk.Context, _ uint64, _ sdk.AccAddress) {
}

func (h Hooks) AfterProposalInactive(_ sdk.Context, _ uint64) {}

func (h Hooks) AfterProposalActive(_ sdk.Context, _ uint64) {}

// staking hooks
func (h Hooks) AfterValidatorCreated(_ sdk.Context, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeValidatorModified(_ sdk.Context, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorRemoved(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBonded(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBeginUnbonding(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationCreated(_ sdk.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationSharesModified(_ sdk.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationRemoved(_ sdk.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeValidatorSlashed(_ sdk.Context, _ sdk.ValAddress, _ sdk.Dec) error {
	return nil
}

// IBC callbacks and transfer handlers

// SendPacket implements the ICS4Wrapper interface from the transfer module. It
// calls the underlying SendPacket function directly to move down the middleware
// stack. Without SendPacket, this module would be skipped, when sending packages
// from the transferKeeper to core IBC.
func (k Keeper) SendPacket(
	ctx sdk.Context,
	channelCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	sequence, err = k.ics4Wrapper.SendPacket(
		ctx,
		channelCap,
		sourcePort,
		sourceChannel,
		timeoutHeight,
		timeoutTimestamp,
		data,
	)
	if err != nil {
		return 0, err
	}
	return sequence, nil
}

// WriteAcknowledgement implements the ICS4Wrapper interface from the transfer module.
// It calls the underlying WriteAcknowledgement function directly to move down the middleware stack.
func (k Keeper) WriteAcknowledgement(ctx sdk.Context, channelCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, channelCap, packet, ack)
}

// GetAppVersion returns the underlying application version.
func (k Keeper) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return k.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}
