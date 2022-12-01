package keeper_test

import (
	"github.com/evmos/evmos/v10/x/inflation/types"
)

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.InflationKeeper.GetParams(suite.ctx)
	expParams := types.DefaultParams()

	suite.Require().Equal(expParams, params)

	err := suite.app.InflationKeeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	newParams := suite.app.InflationKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}
