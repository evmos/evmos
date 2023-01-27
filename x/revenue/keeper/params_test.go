package keeper_test

import "github.com/evmos/evmos/v11/x/revenue/types"

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.RevenueKeeper.GetParams(suite.ctx)
	params.EnableRevenue = true
	suite.Require().Equal(types.DefaultParams(), params)
	params.EnableRevenue = false
	err := suite.app.RevenueKeeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)
	newParams := suite.app.RevenueKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}
