package keeper_test

import "github.com/evmos/evmos/v15/x/incentives/types"

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.IncentivesKeeper.GetParams(suite.ctx)
	suite.Require().Equal(types.DefaultParams(), params)
	params.EnableIncentives = false
	err := suite.app.IncentivesKeeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)
	newParams := suite.app.IncentivesKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}
