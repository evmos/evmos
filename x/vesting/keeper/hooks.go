// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	vestingtypes "github.com/evmos/evmos/v14/x/vesting/types"
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
func (h Hooks) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {
	h.k.AfterProposalSubmission(ctx, proposalID)
}

// AfterProposalSubmission is called after a governance clawback proposal is submitted on chain.
// It adds a store entry for the combination of vesting account and funder address for the time
// the proposal is active in order to prevent manual clawback from the funder, which could overrule
// the community vote.
func (k Keeper) AfterProposalSubmission(_ sdk.Context, _ uint64) {}

// AfterProposalDeposit is a wrapper for calling the Gov AfterProposalDeposit hook on
// the module keeper
func (h Hooks) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) {
	h.k.AfterProposalDeposit(ctx, proposalID, depositorAddr)
}

// AfterProposalDeposit is called after a deposit is made on a governance clawback proposal.
func (k Keeper) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, _ sdk.AccAddress) {
	fmt.Println("Running AfterProposalDeposit")
	proposal, found := k.govKeeper.GetProposal(ctx, proposalID)
	if !found {
		k.Logger(ctx).Error("proposal not found",
			"proposalID", proposalID,
			"hook", "AfterProposalSubmission",
		)
		return
	}

	clawbackProposal, isClawbackProposal, err := getClawbackProposal(proposal)
	if err != nil {
		k.Logger(ctx).Error("failed to get clawback proposal",
			"proposalID", proposalID,
			"hook", "AfterProposalSubmission",
			"error", err,
		)
		return
	}
	if !isClawbackProposal {
		// no-op when proposal is not a clawback proposal
		return
	}

	totalDeposit := sdk.NewCoins(proposal.GetTotalDeposit()...)
	fmt.Println("Total deposit: ", totalDeposit)
	govParams := k.govKeeper.GetParams(ctx)
	minDeposit := sdk.NewCoins(govParams.MinDeposit...)
	fmt.Println("Min deposit: ", minDeposit)
	if totalDeposit.IsAllLT(minDeposit) {
		fmt.Println("Proposal deposit is less than min deposit")
		return
	}

	// TODO: do we need check here if there is already an active proposal?
	// or should we check that in the proposal handler? Probably better there
	vesting := sdk.MustAccAddressFromBech32(clawbackProposal.Address)
	k.SetActiveClawbackProposal(ctx, vesting)
}

// AfterProposalVote is a wrapper for calling the Gov AfterProposalVote hook on
// the module keeper
func (h Hooks) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	h.k.AfterProposalVote(ctx, proposalID, voterAddr)
}

// AfterProposalVote is called after a vote on a governance clawback proposal is cast.
func (k Keeper) AfterProposalVote(_ sdk.Context, _ uint64, _ sdk.AccAddress) {}

// AfterProposalFailedMinDeposit is a wrapper for calling the Gov AfterProposalFailedMinDeposit hook on
// the module keeper
func (h Hooks) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) {
	h.k.AfterProposalFailedMinDeposit(ctx, proposalID)
}

// AfterProposalFailedMinDeposit is called after a governance clawback proposal fails due to
// not meeting the minimum deposit.
func (k Keeper) AfterProposalFailedMinDeposit(_ sdk.Context, _ uint64) {}

// AfterProposalVotingPeriodEnded is a wrapper for calling the Gov AfterProposalVotingPeriodEnded hook on
// the module keeper
func (h Hooks) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {
	h.k.AfterProposalVotingPeriodEnded(ctx, proposalID)
}

// AfterProposalVotingPeriodEnded is called after the voting period of a governance clawback proposal
// has ended.
func (k Keeper) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {
	proposal, found := k.govKeeper.GetProposal(ctx, proposalID)
	if !found {
		k.Logger(ctx).Error("proposal not found",
			"proposalID", proposalID,
			"hook", "AfterProposalVotingPeriodEnded",
		)
		return
	}

	clawbackProposal, isClawbackProposal, err := getClawbackProposal(proposal)
	if err != nil {
		k.Logger(ctx).Error("failed to get clawback proposal",
			"proposalID", proposalID,
			"hook", "AfterProposalVotingPeriodEnded",
			"error", err,
		)
		return
	}
	if !isClawbackProposal {
		// no-op when proposal is not a clawback proposal
		return
	}

	vesting := sdk.MustAccAddressFromBech32(clawbackProposal.Address)
	k.DeleteActiveClawbackProposal(ctx, vesting)
}

// getClawbackProposal checks if the proposal with the given ID is a governance
// clawback proposal.
func getClawbackProposal(proposal govv1.Proposal) (vestingtypes.ClawbackProposal, bool, error) {
	// TODO: do we have to check here or is this hook only called for clawback governance proposals?
	msgs, err := proposal.GetMsgs()
	if err != nil {
		return vestingtypes.ClawbackProposal{}, false, err
	}
	if len(msgs) == 0 {
		return vestingtypes.ClawbackProposal{}, false, errors.New("proposal has no messages")
	}

	msgContent, ok := msgs[0].(*govv1.MsgExecLegacyContent)
	if !ok {
		return vestingtypes.ClawbackProposal{}, false, errors.New("failed to cast msg to MsgExecLegacyContent")
	}

	clawbackProposal, ok := msgContent.Content.GetCachedValue().(*vestingtypes.ClawbackProposal)
	if !ok {
		// NOTE: no need to return an error here, it's expected that proposals are not a clawback proposal
		return vestingtypes.ClawbackProposal{}, false, nil
	}

	return *clawbackProposal, true, nil
}
