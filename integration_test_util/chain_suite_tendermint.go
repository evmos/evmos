package integration_test_util

//goland:noinspection SpellCheckingInspection

// HasTendermint indicate if the integration chain has Tendermint enabled.
func (suite *ChainIntegrationTestSuite) HasTendermint() bool {
	return !suite.TestConfig.DisableTendermint && suite.TendermintApp != nil
}

// EnsureTendermint trigger test failure immediately if Tendermint is not enabled on integration chain.
func (suite *ChainIntegrationTestSuite) EnsureTendermint() {
	if !suite.HasTendermint() {
		suite.Require().FailNow("tendermint node must be initialized")
	}
}
