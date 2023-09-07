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
func (k Keeper) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {
	fmt.Printf("AfterProposalSubmission: %d\n", proposalID)
	k.Logger(ctx).Info(
		"Running AfterProposalSubmission hook",
		"proposalID", proposalID,
	)
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
	// or should we check that in the proposal handler? Probably better there
	k.SetActiveClawbackProposal(ctx, vesting, funder)
}

// AfterProposalDeposit is a wrapper for calling the Gov AfterProposalDeposit hook on
// the module keeper
func (h Hooks) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) {
	h.k.AfterProposalDeposit(ctx, proposalID, depositorAddr)
}

// AfterProposalDeposit is called after a deposit is made on a governance clawback proposal.
func (k Keeper) AfterProposalDeposit(_ sdk.Context, _ uint64, _ sdk.AccAddress) {}

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

// AfterProposalVotingPeriodEnded is a wrapper for calling the Gov AfterProposalVotingPeriodEnded hook on
// the module keeper
func (h Hooks) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {
	h.k.AfterProposalVotingPeriodEnded(ctx, proposalID)
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
func (k Keeper) getClawbackProposal(ctx sdk.Context, proposalID uint64) (govv1.Proposal, bool, error) {
	proposal, found := k.govKeeper.GetProposal(ctx, proposalID)
	if !found {
		return govv1.Proposal{}, false, errors.New("proposal not found")
	}

	// TODO: do we have to check here or is this hook only called for clawback governance proposals?
	// FIXME: check proposal type
	return proposal, false, nil
}

// getAddressAndFunder returns the vesting account address and funder address for the given
// governance clawback proposal.
func (k Keeper) getAddressAndFunder(ctx sdk.Context, proposal govv1.Proposal) (sdk.AccAddress, sdk.AccAddress, error) {
	msgs, err := proposal.GetMsgs()
	if err != nil {
		return sdk.AccAddress{}, sdk.AccAddress{}, err
	}
	if len(msgs) == 0 {
		return sdk.AccAddress{}, sdk.AccAddress{}, errors.New("proposal has no messages")
	}

	fmt.Printf("messages: %v\n", msgs)
	msgContent := msgs[0].(*govv1.MsgExecLegacyContent)
	clawbackProposal, ok := msgContent.Content.GetCachedValue().(*vestingtypes.ClawbackProposal)
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
