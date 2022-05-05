package keeper_test

import "github.com/tharsis/evmos/v4/x/fees/types"

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.FeesKeeper.GetParams(suite.ctx)
	params.EnableFees = false
	suite.Require().Equal(types.DefaultParams(), params)
	params.EnableFees = true
	suite.app.FeesKeeper.SetParams(suite.ctx, params)
	newParams := suite.app.FeesKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}
