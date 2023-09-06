package keeper

import (
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	vestingtypes "github.com/evmos/evmos/v14/x/vesting/types"
)

var _ govtypes.GovHooks = Keeper{}

// AfterProposalSubmission is called after a governance clawback proposal is submitted on chain.
// It adds a store entry for the combination of vesting account and funder address for the time
// the proposal is active in order to prevent manual clawback from the funder, which could overrule
// the community vote.
func (k Keeper) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {
	proposal, isClawbackProposal, err := k.getClawbackProposal(ctx, proposalID)
	if err != nil {
		k.Logger(ctx).Error("failed to check proposal",
			"proposalID", proposalID,
			"hook", "AfterProposalFailedMinDeposit",
		)
		return
	}
	if !isClawbackProposal {
		// no-op when proposal is not a clawback proposal
		return
	}

	vesting, funder, err := k.getAddressAndFunder(ctx, proposal)
	if err != nil {
		k.Logger(ctx).Error(
			"failed to get clawback vesting account",
			"proposalID", proposalID,
			"hook", "AfterProposalFailedMinDeposit",
		)
		return
	}

	// TODO: do we need check here if there is already an active proposal?
	// or should we check that in the proposal handler?
	k.SetActiveClawbackProposal(ctx, vesting, funder)
}

// AfterProposalDeposit is called after a deposit is made on a governance clawback proposal.
func (k Keeper) AfterProposalDeposit(_ sdk.Context, _ uint64, _ sdk.AccAddress) {}

// AfterProposalVote is called after a vote on a governance clawback proposal is cast.
func (k Keeper) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {}

// AfterProposalFailedMinDeposit is called after a governance clawback proposal fails due to
// not meeting the minimum deposit.
func (k Keeper) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) {
	proposal, isClawbackProposal, err := k.getClawbackProposal(ctx, proposalID)
	if err != nil {
		k.Logger(ctx).Error("failed to check proposal",
			"proposalID", proposalID,
			"hook", "AfterProposalFailedMinDeposit",
		)
		return
	}
	if !isClawbackProposal {
		// no-op when proposal is not a clawback proposal
		return
	}

	vesting, funder, err := k.getAddressAndFunder(ctx, proposal)
	if err != nil {
		k.Logger(ctx).Error(
			"failed to get clawback vesting account",
			"proposalID", proposalID,
			"hook", "AfterProposalFailedMinDeposit",
			"error", err,
		)
		return
	}

	k.DeleteActiveClawbackProposal(ctx, vesting, funder)
}

// AfterProposalVotingPeriodEnded is called after the voting period of a governance clawback proposal
// has ended.
func (k Keeper) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {
	proposal, isClawbackProposal, err := k.getClawbackProposal(ctx, proposalID)
	if err != nil {
		k.Logger(ctx).Error("failed to check proposal",
			"proposalID", proposalID,
			"hook", "AfterProposalFailedMinDeposit",
			"error", err,
		)
		return
	}
	if !isClawbackProposal {
		// no-op when proposal is not a clawback proposal
		return
	}

	vesting, funder, err := k.getAddressAndFunder(ctx, proposal)
	if err != nil {
		k.Logger(ctx).Error(
			"failed to get clawback vesting account",
			"proposalID", proposalID,
			"hook", "AfterProposalFailedMinDeposit",
			"error", err,
		)
		return
	}

	k.DeleteActiveClawbackProposal(ctx, vesting, funder)
}

// getClawbackProposal checks if the proposal with the given ID is a governance
// clawback proposal.
func (k Keeper) getClawbackProposal(ctx sdk.Context, proposalID uint64) (govv1beta1.Proposal, bool, error) {
	proposal, found := k.govKeeper.GetProposal(ctx, proposalID)
	if !found {
		return govv1beta1.Proposal{}, false, errors.New("proposal not found")
	}

	// TODO: do we have to check here or is this hook only called for clawback governance proposals?
	return proposal, proposal.ProposalType() == vestingtypes.ProposalTypeClawback, nil
}

// getAddressAndFunder returns the vesting account address and funder address for the given
// governance clawback proposal.
func (k Keeper) getAddressAndFunder(ctx sdk.Context, proposal govv1beta1.Proposal) (sdk.AccAddress, sdk.AccAddress, error) {
	content := proposal.GetContent()
	clawbackProposal, ok := content.(*vestingtypes.ClawbackProposal)
	if !ok {
		return sdk.AccAddress{}, sdk.AccAddress{}, errors.New("proposal content is not a clawback proposal")
	}

	// TODO: do we need to check error here? Should only be possible to store a valid address right?
	vestingAccAddr, err := sdk.AccAddressFromBech32(clawbackProposal.Address)
	if err != nil {
		return sdk.AccAddress{}, sdk.AccAddress{}, err
	}

	vesting, err := k.GetClawbackVestingAccount(ctx, vestingAccAddr)
	if err != nil {
		return sdk.AccAddress{}, sdk.AccAddress{}, err
	}

	// TODO: do we need to check error here? Should only be possible to store a valid address right?
	funder, err := sdk.AccAddressFromBech32(vesting.FunderAddress)
	if err != nil {
		return sdk.AccAddress{}, sdk.AccAddress{}, err
	}

	return vestingAccAddr, funder, nil
}
