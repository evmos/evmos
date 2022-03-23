package keeper_test

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.WithdrawKeeper.GetParams(suite.ctx)
	params.EnableWithdraw = false
	suite.app.WithdrawKeeper.SetParams(suite.ctx, params)
	newParams := suite.app.WithdrawKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}
