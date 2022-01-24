package keeper_test

import (
	"github.com/tharsis/evmos/x/claims/types"
)

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.ClaimKeeper.GetParams(suite.ctx)
	expParams := types.DefaultParams()
	expParams.AirdropStartTime = suite.ctx.BlockTime()
	suite.Require().Equal(expParams, params)
	params.EnableClaim = false
	suite.app.ClaimKeeper.SetParams(suite.ctx, params)
	newParams := suite.app.ClaimKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}
