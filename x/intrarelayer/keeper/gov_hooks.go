package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

var _ govtypes.GovHooks = &Keeper{}

// AfterProposalSubmission performs a no-op
func (k Keeper) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {}

// AfterProposalDeposit hook overrides the voting period for the
// RegisterTokenPairProposal to the value defined on the intrarelayer module
// parameters.
func (k Keeper) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, _ sdk.AccAddress) {
	// fetch the original voting period from gov params
	votingPeriod := k.govKeeper.GetVotingParams(ctx).VotingPeriod
	// get the new voting period
	newVotingPeriod := k.GetVotingPeriod(ctx, types.ProposalTypeRegisterCoin)

	// perform a no-op if voting periods are equal
	if newVotingPeriod == votingPeriod {
		return
	}

	// get proposal
	proposal, found := k.govKeeper.GetProposal(ctx, proposalID)
	if !found {
		return
	}

	// check if the proposal is on voting period
	if proposal.Status != govtypes.StatusVotingPeriod {
		return
	}

	content := proposal.GetContent()

	// check if proposal content and type matches the given type
	if content.ProposalType() != types.ProposalTypeRegisterCoin && content.ProposalType() != types.ProposalTypeRegisterERC20 {
		return
	}

	switch content.(type) {
	case *types.RegisterCoinProposal, *types.RegisterERC20Proposal:
		// valid proposal types
	default:
		return
	}

	originalEndTime := proposal.VotingEndTime
	proposal.VotingEndTime = proposal.VotingStartTime.Add(newVotingPeriod)

	// remove old proposal from the queue with old voting end time
	k.govKeeper.RemoveFromActiveProposalQueue(ctx, proposalID, originalEndTime)
	// reinsert the proposal to the queue with the updated voting end time
	k.govKeeper.InsertActiveProposalQueue(ctx, proposalID, proposal.VotingEndTime)
	// update the proposal
	k.govKeeper.SetProposal(ctx, proposal)

	k.govKeeper.Logger(ctx).Info("proposal voting end time updated", "id", proposalID, "endtime", proposal.VotingEndTime.String())
}

// AfterProposalVote performs a no-op
func (k Keeper) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {}

// AfterProposalFailedMinDeposit performs a no-op
func (k Keeper) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) {}

// AfterProposalVotingPeriodEnded performs a no-op
func (k Keeper) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {}

// GetVotingPeriod implements the ProposalHook interface
func (k Keeper) GetVotingPeriod(ctx sdk.Context, proposalType string) time.Duration {
	params := k.GetParams(ctx)

	switch proposalType {
	case types.ProposalTypeRegisterCoin:
		return params.TokenPairVotingPeriod
	case types.ProposalTypeRegisterERC20:
		return params.TokenPairVotingPeriod
	default:
		vp := k.govKeeper.GetVotingParams(ctx)
		return vp.VotingPeriod
	}
}
