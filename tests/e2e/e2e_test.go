package e2e

func (s *IntegrationTestSuite) TestUpgrade() {
	s.runInitialNode()
	s.proposeUpgrade()
	s.depositToProposal()
	s.voteForProposal()
	s.upgrade()
	s.T().Logf("SUCCESS")
}
