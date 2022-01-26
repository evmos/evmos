package keeper_test

import (
	"github.com/tharsis/evmos/x/inflation/types"
)

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.InflationKeeper.GetParams(suite.ctx)
	expParams := types.DefaultParams()

	suite.Require().Equal(expParams, params)

	params.EpochsPerPeriod = 700
	suite.app.InflationKeeper.SetParams(suite.ctx, params)
	newParams := suite.app.InflationKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}
