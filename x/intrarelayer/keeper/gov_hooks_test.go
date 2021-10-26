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
	votingPeriod := suite.app.IntrarelayerKeeper.GetVotingPeriod(suite.ctx, "")
	expPeriod := votingPeriod + 1<<45

	proposalID := uint64(1)

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
		// TODO: Different Status test
		// {
		// 	"don't override voting period (different status)",
		// 	func() {
		// 		params := types.Params{TokenPairVotingPeriod: expPeriod}
		// 		suite.app.IntrarelayerKeeper.SetParams(suite.ctx, params)

		// 		pair := types.NewTokenPair(tests.GenerateAddress(), "coin", true)

		// 		content := types.NewRegisterTokenPairProposal("title", "desc", pair)
		// 		proposal, err := govtypes.NewProposal(content, proposalID, time.Now().UTC(), time.Now().UTC())
		// 		suite.Require().NoError(err)

		// 		proposal.Status = govtypes.ProposalStatus(0)
		// 		suite.app.GovKeeper.SetProposal(suite.ctx, proposal)

		// 		if proposal.Status != govtypes.StatusVotingPeriod {
		// 			fmt.Println()
		// 		}

		// 		fmt.Printf("\npropsal.Status: %s\n", proposal.Status)
		// 		fmt.Printf("govtypes.StatusVotingPeriod: %s\n", govtypes.StatusVotingPeriod)
		// 	},
		// 	true,
		// },
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()
			suite.app.GovKeeper.AfterProposalDeposit(suite.ctx, proposalID, sdk.AccAddress{})
			newVotingPeriod := suite.app.IntrarelayerKeeper.GetVotingPeriod(suite.ctx, types.ProposalTypeRegisterTokenPair)

			if tc.noOp {
				suite.Require().Equal(votingPeriod, newVotingPeriod)
			} else {
				suite.Require().Equal(expPeriod, newVotingPeriod)
			}
		})
	}
}
