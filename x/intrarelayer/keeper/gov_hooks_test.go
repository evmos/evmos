package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

func (suite *KeeperTestSuite) TestProposalNoOps() {
	suite.app.GovKeeper.AfterProposalSubmission(suite.ctx, uint64(0))
	suite.app.GovKeeper.AfterProposalVote(suite.ctx, uint64(0), sdk.AccAddress{})
	suite.app.GovKeeper.AfterProposalFailedMinDeposit(suite.ctx, uint64(0))
	suite.app.GovKeeper.AfterProposalVotingPeriodEnded(suite.ctx, uint64(0))
}

func (suite *KeeperTestSuite) TestAfterProposalDeposit() {
	// proposal := suite.app.IntrarelayerKeeper.GetVotingPeriod(suite.ctx, types.ProposalTypeRegisterTokenPair)
	votingPeriod := suite.app.IntrarelayerKeeper.GetVotingPeriod(suite.ctx, "")
	expPeriod := votingPeriod + 1<<45

	testCases := []struct {
		name     string
		malleate func()
		noOp     bool
	}{
		{
			"override voting period",
			func() {
				params := types.Params{TokenPairVotingPeriod: expPeriod}
				suite.app.IntrarelayerKeeper.SetParams(suite.ctx, params)
			},
			false,
		},
		{
			"don't override voting period (same duration)",
			func() {
				params := types.Params{TokenPairVotingPeriod: votingPeriod}
				suite.app.IntrarelayerKeeper.SetParams(suite.ctx, params)
			},
			true,
		},
		// {
		// 	"don't override voting period (different status)",
		// 	func() {

		// 	},
		// 	true,
		// },
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()
			suite.app.GovKeeper.AfterProposalDeposit(suite.ctx, uint64(0), sdk.AccAddress{})

			newVotingPeriod := suite.app.IntrarelayerKeeper.GetVotingPeriod(suite.ctx, types.ProposalTypeRegisterTokenPair)
			if tc.noOp {
				suite.Require().Equal(votingPeriod, newVotingPeriod)
			} else {
				suite.Require().Equal(expPeriod, newVotingPeriod)
			}
		})
	}
}
