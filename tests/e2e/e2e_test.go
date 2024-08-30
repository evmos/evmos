package e2e

import (
	"context"
)

// TestUpgrade tests if an Evmos node can be upgraded from one version to another.
// It iterates through the list of scheduled upgrades, that are defined using the input
// arguments to the make command. The function then submits a proposal to upgrade the chain,
// and finally upgrades the chain.
// If the chain can be restarted after the upgrade(s), the test passes.
func (s *IntegrationTestSuite) TestUpgrade() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// NOTE: we initialize the current version, which is then updated in the upgrade function
	s.upgradeManager.CurrentVersion = s.upgradeParams.Versions[0].ImageTag
	for idx, version := range s.upgradeParams.Versions {
		if idx == 0 {
			// start initial node
			s.runInitialNode(version)
			continue
		}

		// wait one block to execute the txs
		err := s.upgradeManager.WaitNBlocks(ctx, 1)
		s.Require().NoError(err)
		s.T().Logf("(upgrade %d): UPGRADING TO %s WITH PROPOSAL NAME %s", idx, version.ImageTag, version.UpgradeName)
		s.proposeUpgrade(version.UpgradeName, version.ImageTag)

		s.Require().NoError(s.upgradeManager.WaitNBlocks(ctx, 1), "failed to wait for block")

		s.voteForProposal(idx)
		s.upgrade(version)
	}
	s.T().Logf("SUCCESS")
}
