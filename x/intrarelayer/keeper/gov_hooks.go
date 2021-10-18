package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

var (
	_ govtypes.GovHooks          = &GovHooks{}
	_ types.VotingPeriodModifier = &Keeper{}
)

// GovHooks
type GovHooks struct {
	modifier  types.VotingPeriodModifier
	govKeeper types.GovKeeper
}

func NewGovHooks(m types.VotingPeriodModifier, gk types.GovKeeper) GovHooks {
	return GovHooks{
		modifier:  m,
		govKeeper: gk,
	}
}

// AfterProposalVotingPeriodEnded performs a no-op
func (h GovHooks) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {}

// AfterProposalDeposit hook overrides the voting period for the RegisterTokenPairProposal to the
// value defined on the intrarelayer module parameters.
func (h GovHooks) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, _ sdk.AccAddress) {
	votingPeriod := h.govKeeper.GetVotingParams(ctx).VotingPeriod
	newVotingPeriod := h.modifier.GetVotingPeriod(ctx, types.ProposalTypeRegisterTokenPair)

	// perform a no-op if voting periods are equal
	if newVotingPeriod == votingPeriod {
		return
	}

	proposal, found := h.govKeeper.GetProposal(ctx, proposalID)
	if !found {
		return
	}

	// check if the proposal is on voting period
	if proposal.Status != govtypes.StatusVotingPeriod {
		return
	}

	content := proposal.GetContent()

	// check if proposal content and type matches the given type
	if content.ProposalType() != types.ProposalTypeRegisterTokenPair {
		return
	}

	if _, ok := content.(*types.RegisterTokenPairProposal); !ok {
		return
	}

	originalEndTime := proposal.VotingEndTime
	proposal.VotingEndTime = proposal.VotingStartTime.Add(newVotingPeriod)

	// remove old proposal with old voting end time and reinsert it with the updated voting end time
	h.govKeeper.RemoveFromActiveProposalQueue(ctx, proposalID, originalEndTime)
	h.govKeeper.InsertActiveProposalQueue(ctx, proposalID, proposal.VotingEndTime)

	h.govKeeper.Logger(ctx).Info("proposal voting end time updated", "id", proposalID, "endtime", proposal.VotingEndTime.String())
}

// AfterProposalVotingPeriodEnded performs a no-op
func (h GovHooks) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {}

// AfterProposalFailedMinDeposit performs a no-op
func (h GovHooks) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) {}

// AfterProposalVotingPeriodEnded performs a no-op
func (h GovHooks) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {}

// GetVotingPeriod implements the ProposalHook interface
func (k Keeper) GetVotingPeriod(ctx sdk.Context, proposalType string) time.Duration {
	params := k.GetParams(ctx)

	switch proposalType {
	case types.ProposalTypeRegisterTokenPair:
		return params.TokenPairVotingPeriod
	default:
		return 0
	}
}
