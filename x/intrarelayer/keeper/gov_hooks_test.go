package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"
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
			"don't override voting period (same duration)",
			func() {
				params := types.Params{TokenPairVotingPeriod: votingPeriod}
				suite.app.IntrarelayerKeeper.SetParams(suite.ctx, params)

				content := types.NewRegisterERC20Proposal("title", "desc", common.Address{}.String())
				proposal, err := govtypes.NewProposal(content, proposalID, time.Now().UTC(), time.Now().UTC())
				suite.Require().NoError(err)

				suite.app.GovKeeper.ActivateVotingPeriod(suite.ctx, proposal)
			},
			false,
		},
		{
			"don't override voting period (different status)",
			func() {
				params := types.Params{TokenPairVotingPeriod: votingPeriod}
				suite.app.IntrarelayerKeeper.SetParams(suite.ctx, params)

				content := types.NewRegisterERC20Proposal("title", "desc", common.Address{}.String())
				proposal, err := govtypes.NewProposal(content, proposalID, time.Now().UTC(), time.Now().UTC())
				suite.Require().NoError(err)

				//activate proposal
				suite.app.GovKeeper.ActivateVotingPeriod(suite.ctx, proposal)

				// override proposal status
				proposal, _ = suite.app.GovKeeper.GetProposal(suite.ctx, proposalID)
				proposal.Status = govtypes.ProposalStatus(0)
				suite.app.GovKeeper.SetProposal(suite.ctx, proposal)

				// update params after proposal creation
				params = types.Params{TokenPairVotingPeriod: expPeriod}
				suite.app.IntrarelayerKeeper.SetParams(suite.ctx, params)
			},
			false,
		},
		{
			"override voting period",
			func() {
				params := types.Params{TokenPairVotingPeriod: votingPeriod}
				suite.app.IntrarelayerKeeper.SetParams(suite.ctx, params)

				content := types.NewRegisterERC20Proposal("title", "desc", common.Address{}.String())
				proposal, err := govtypes.NewProposal(content, proposalID, time.Now().UTC(), time.Now().UTC())
				suite.Require().NoError(err)

				suite.app.GovKeeper.ActivateVotingPeriod(suite.ctx, proposal)

				// update params after proposal creation
				params = types.Params{TokenPairVotingPeriod: expPeriod}
				suite.app.IntrarelayerKeeper.SetParams(suite.ctx, params)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()
			suite.app.GovKeeper.AfterProposalDeposit(suite.ctx, proposalID, sdk.AccAddress{})
			newVotingPeriod := suite.app.IntrarelayerKeeper.GetVotingPeriod(suite.ctx, types.ProposalTypeRegisterCoin)

			proposal, ok := suite.app.GovKeeper.GetProposal(suite.ctx, proposalID)

			if tc.noOp {
				suite.Require().True(ok)
				// Proposal time was updated
				suite.Require().Equal(proposal.VotingEndTime, proposal.VotingStartTime.Add(newVotingPeriod))
			} else {
				suite.Require().True(ok)
				// Proposal time was not updated
				suite.Require().Equal(proposal.VotingEndTime, proposal.VotingStartTime.Add(votingPeriod))
			}
		})
	}
}
