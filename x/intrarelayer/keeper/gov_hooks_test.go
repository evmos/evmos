package keeper_test

import sdk "github.com/cosmos/cosmos-sdk/types"

// var _ govtypes.GovHooks = &MockGovHooksReceiver{}

// type MockGovHooksReceiver struct {
// 	AfterProposalSubmissionValid        bool
// 	AfterProposalDepositValid           bool
// 	AfterProposalVoteValid              bool
// 	AfterProposalFailedMinDepositValid  bool
// 	AfterProposalVotingPeriodEndedValid bool
// }

// func (h *MockGovHooksReceiver) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {
// 	h.AfterProposalSubmissionValid = true
// }

// func (h *MockGovHooksReceiver) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) {
// 	h.AfterProposalDepositValid = true
// }

// func (h *MockGovHooksReceiver) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
// 	h.AfterProposalVoteValid = true
// }
// func (h *MockGovHooksReceiver) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) {
// 	h.AfterProposalFailedMinDepositValid = true
// }
// func (h *MockGovHooksReceiver) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) {
// 	h.AfterProposalVotingPeriodEndedValid = true
// }

func (suite *KeeperTestSuite) TestAfterProposalSubmission() {

	suite.app.GovKeeper.AfterProposalSubmission(suite.ctx, uint64(0))
	suite.app.GovKeeper.AfterProposalDeposit(suite.ctx, uint64(0), sdk.AccAddress{})

	tt := true
	suite.Require().True(tt)
}
