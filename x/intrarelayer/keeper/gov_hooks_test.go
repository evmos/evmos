package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

func (suite *KeeperTestSuite) TestAfterProposalDeposit() {
	testCases := []struct {
		name     string
		malleate func()
		noOp     bool
	}{
		{"override voting period", func() {}, false},
		{"don't voting period (same duration)", func() {}, false},
		{"don't voting period (different duration)", func() {}, false},
	}
}

func (suite *KeeperTestSuite) TestAfterProposalSubmission() {
	suite.app.GovKeeper.AfterProposalSubmission(suite.ctx, uint64(0))
	suite.app.GovKeeper.AfterProposalDeposit(suite.ctx, uint64(0), sdk.AccAddress{})
}

func (suite *KeeperTestSuite) TestGetVotingPeriod() {
	period := suite.app.IntrarelayerKeeper.GetVotingPeriod(suite.ctx, types.ProposalTypeRegisterTokenPair)
}
