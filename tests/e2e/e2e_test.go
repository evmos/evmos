package e2e

import "github.com/evmos/evmos/v11/tests/e2e/upgrade"

// TestUpgrade tests if an Evmos node can be upgraded from one version to another.
// It iterates through the list of scheduled upgrades, that are defined using the input
// arguments to the make command. The function then submits a proposal to upgrade the chain,
// and finally upgrades the chain.
// If the chain can be restarted after the upgrade(s), the test passes.
func (s *IntegrationTestSuite) TestUpgrade() {
	for idx, version := range s.upgradeParams.Versions {
		if idx == 0 {
			// start initial node
			s.runInitialNode(version, registryDockerFile)
			continue
		}
		s.T().Logf("(upgrade %d): UPGRADING TO %s WITH PROPOSAL NAME %s", idx, version.ImageTag, version.UpgradeName)
		s.proposeUpgrade(version.UpgradeName, version.ImageTag)
		s.voteForProposal(idx)
		s.upgrade(version.ImageName, version.ImageTag)
	}
	s.T().Logf("SUCCESS")
}

func (s *IntegrationTestSuite) TestCLITxs() {
	mainBranch := upgrade.VersionConfig{
		ImageTag:  "main",
		ImageName: "evmos",
	}

	newVersion := upgrade.VersionConfig{
		UpgradeName: "v11.0.0",
		ImageTag:    "v11.0.0",
		ImageName:   "evmos",
	}

	s.runInitialNode(mainBranch, repoDockerFile)

	s.proposeUpgrade(newVersion.UpgradeName, newVersion.ImageTag)
	s.voteForProposal(1)
	s.upgrade(newVersion.ImageName, newVersion.ImageTag)
}
