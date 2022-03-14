package keeper_test

import (
	"github.com/tharsis/evmos/v2/x/withdraw/types"
)

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.WithdrawKeeper.GetParams(suite.ctx)
	expParams := types.DefaultParams()
	params.EnableWithdraw = false
	expParams.EnableWithdraw = false

	suite.app.WithdrawKeeper.SetParams(suite.ctx, params)
	newParams := suite.app.WithdrawKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}
