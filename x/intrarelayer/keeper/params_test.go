package keeper_test

import "github.com/tharsis/evmos/x/intrarelayer/types"

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.IntrarelayerKeeper.GetParams(suite.ctx)
	suite.Require().Equal(types.DefaultParams(), params)
	params.EnableIntrarelayer = false
	suite.app.IntrarelayerKeeper.SetParams(suite.ctx, params)
	newParams := suite.app.IntrarelayerKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}
