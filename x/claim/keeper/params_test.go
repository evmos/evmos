package keeper_test

import (
	"github.com/tharsis/evmos/x/claim/types"
)

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.ClaimKeeper.GetParams(suite.ctx)
	suite.Require().Equal(types.DefaultParams(), params)
	params.EnableClaim = false
	suite.app.ClaimKeeper.SetParams(suite.ctx, params)
	newParams := suite.app.ClaimKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}
