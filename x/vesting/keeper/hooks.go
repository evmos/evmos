// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	vestingtypes "github.com/evmos/evmos/v16/x/vesting/types"
)

var _ govtypes.GovHooks = Hooks{}

// Hooks wrapper struct for the vesting keeper
type Hooks struct {
	k Keeper
}

// Hooks returns the wrapper struct for vesting hooks
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// AfterProposalSubmission is a wrapper for calling the Gov AfterProposalSubmission hook on
// the module keeper
func (h Hooks) AfterProposalSubmission(ctx context.Context, proposalID uint64) error {
	return h.k.AfterProposalSubmission(ctx, proposalID)
}

// AfterProposalSubmission is called after a governance clawback proposal is submitted on chain.
// It adds a store entry for the vesting account for the time the proposal is active
// in order to prevent manual clawback from the funder, which could overrule the community vote.
func (k Keeper) AfterProposalSubmission(_ context.Context, _ uint64) error {
	return nil
}

// AfterProposalDeposit is a wrapper for calling the Gov AfterProposalDeposit hook on
// the module keeper
func (h Hooks) AfterProposalDeposit(ctx context.Context, proposalID uint64, depositorAddr sdk.AccAddress) error {
	return h.k.AfterProposalDeposit(ctx, proposalID, depositorAddr)
}

// AfterProposalDeposit is called after a deposit is made on a governance clawback proposal.
func (k Keeper) AfterProposalDeposit(c context.Context, proposalID uint64, _ sdk.AccAddress) error {
	ctx := sdk.UnwrapSDKContext(c)
	proposal, err := k.govKeeper.Proposals.Get(ctx, proposalID)
	if err != nil {
		k.Logger(ctx).Error("proposal not found",
			"proposalID", proposalID,
			"hook", "AfterProposalSubmission",
		)
		return err
	}

	clawbackProposals, err := getClawbackProposals(proposal)
	if err != nil {
		k.Logger(ctx).Error("failed to get clawback proposal",
			"proposalID", proposalID,
			"hook", "AfterProposalSubmission",
			"error", err,
		)
		return err
	}
	if len(clawbackProposals) == 0 {
		// no-op when proposal does not contain a clawback proposal
		return nil
	}

	govParams, err := k.govKeeper.Params.Get(ctx)
	if err != nil {
		k.Logger(ctx).Error("failed to get gov params",
			"proposalID", proposalID,
			"hook", "AfterProposalSubmission",
			"error", err,
		)
		return err
	}
	minDeposit := sdk.NewCoins(govParams.MinDeposit...)
	totalDeposit := sdk.NewCoins(proposal.GetTotalDeposit()...)
	if totalDeposit.IsAllLT(minDeposit) {
		return nil
	}

	for _, clawbackProposal := range clawbackProposals {
		vestingAccAddr := sdk.MustAccAddressFromBech32(clawbackProposal.Address)
		k.SetActiveClawbackProposal(ctx, vestingAccAddr)
	}
	return nil
}

// AfterProposalVote is a wrapper for calling the Gov AfterProposalVote hook on
// the module keeper
func (h Hooks) AfterProposalVote(ctx context.Context, proposalID uint64, voterAddr sdk.AccAddress) error {
	return h.k.AfterProposalVote(ctx, proposalID, voterAddr)
}

// AfterProposalVote is called after a vote on a governance clawback proposal is cast.
func (k Keeper) AfterProposalVote(_ context.Context, _ uint64, _ sdk.AccAddress) error {
	return nil
}

// AfterProposalFailedMinDeposit is a wrapper for calling the Gov AfterProposalFailedMinDeposit hook on
// the module keeper
func (h Hooks) AfterProposalFailedMinDeposit(ctx context.Context, proposalID uint64) error {
	return h.k.AfterProposalFailedMinDeposit(ctx, proposalID)
}

// AfterProposalFailedMinDeposit is called after a governance clawback proposal fails due to
// not meeting the minimum deposit.
func (k Keeper) AfterProposalFailedMinDeposit(_ context.Context, _ uint64) error {
	return nil
}

// AfterProposalVotingPeriodEnded is a wrapper for calling the Gov AfterProposalVotingPeriodEnded hook on
// the module keeper
func (h Hooks) AfterProposalVotingPeriodEnded(ctx context.Context, proposalID uint64) error {
	return h.k.AfterProposalVotingPeriodEnded(ctx, proposalID)
}

// AfterProposalVotingPeriodEnded is called after the voting period of a governance clawback proposal
// has ended.
func (k Keeper) AfterProposalVotingPeriodEnded(c context.Context, proposalID uint64) error {
	ctx := sdk.UnwrapSDKContext(c)
	proposal, err := k.govKeeper.Proposals.Get(ctx, proposalID)
	if err != nil {
		k.Logger(ctx).Error("proposal not found",
			"proposalID", proposalID,
			"hook", "AfterProposalVotingPeriodEnded",
		)
		return err
	}

	clawbackProposals, err := getClawbackProposals(proposal)
	if err != nil {
		k.Logger(ctx).Error("failed to get clawback proposal",
			"proposalID", proposalID,
			"hook", "AfterProposalVotingPeriodEnded",
			"error", err,
		)
		return err
	}
	if len(clawbackProposals) == 0 {
		// no-op when proposal content does not contain a clawback proposal
		return nil
	}

	for _, clawbackProposal := range clawbackProposals {
		vestingAccAddr := sdk.MustAccAddressFromBech32(clawbackProposal.Address)
		k.DeleteActiveClawbackProposal(ctx, vestingAccAddr)
	}
	return nil
}

// getClawbackProposals checks if the proposal with the given ID is a governance
// clawback proposal.
func getClawbackProposals(proposal govv1.Proposal) ([]vestingtypes.ClawbackProposal, error) {
	msgs, err := proposal.GetMsgs()
	if err != nil {
		return []vestingtypes.ClawbackProposal{}, err
	}
	if len(msgs) == 0 {
		return []vestingtypes.ClawbackProposal{}, errors.New("proposal has no messages")
	}

	clawbackProposals := make([]vestingtypes.ClawbackProposal, 0, len(msgs))
	for _, msg := range msgs {
		msgContent, ok := msg.(*govv1.MsgExecLegacyContent)
		if !ok {
			continue
		}

		clawbackProposal, ok := msgContent.Content.GetCachedValue().(*vestingtypes.ClawbackProposal)
		if !ok {
			continue
		}

		clawbackProposals = append(clawbackProposals, *clawbackProposal)
	}

	// NOTE: no need to return an error here, it's expected that most proposals do not contain a clawback proposal
	return clawbackProposals, nil
}
