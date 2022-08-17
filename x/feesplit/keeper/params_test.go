package keeper_test

import "github.com/evmos/evmos/v8/x/feesplit/types"

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.FeesplitKeeper.GetParams(suite.ctx)
	params.EnableFeeSplit = true
	suite.Require().Equal(types.DefaultParams(), params)
	params.EnableFeeSplit = false
	suite.app.FeesplitKeeper.SetParams(suite.ctx, params)
	newParams := suite.app.FeesplitKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}
