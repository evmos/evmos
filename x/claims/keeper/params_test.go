package keeper_test

import (
	"github.com/tharsis/evmos/v5/x/claims/types"
)

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
	expParams := types.DefaultParams()
	expParams.AirdropStartTime = suite.ctx.BlockTime()
	suite.Require().Equal(expParams, params)
	params.EnableClaims = false
	suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
	newParams := suite.app.ClaimsKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}
