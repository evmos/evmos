package keeper_test

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.RecoveryKeeper.GetParams(suite.ctx)
	params.EnableRecovery = false
	suite.app.RecoveryKeeper.SetParams(suite.ctx, params)
	newParams := suite.app.RecoveryKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}
