package e2e

// TestUpgrade tests if an Evmos node can be upgraded from one version to another.
// It iterates through the list of scheduled upgrades, that are defined using the input
// arguments to the make command. The function then submits a proposal to upgrade the chain,
// and finally upgrades the chain.
// If the chain can be restarted after the upgrade(s), the test passes.
func (s *IntegrationTestSuite) TestUpgrade() {
	for idx, version := range s.upgradeParams.Versions {
		if idx == 0 {
			// start initial node
			s.runInitialNode(version.tag)
			continue
		}
		s.T().Logf("(upgrade %d): UPGRADING TO %s WITH PROPOSAL NAME %s", idx, version.tag, version.name)
		s.proposeUpgrade(version.name, version.tag)
		s.voteForProposal(idx)
		s.upgrade(version.repo, version.tag)
	}
	s.T().Logf("SUCCESS")
}
