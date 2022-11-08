package e2e

import (
	"os"
	"strconv"
)

func (s *IntegrationTestSuite) loadUpgradeParams() {
	var err error

	initialV := os.Getenv("INITIAL_VERSION")
	if initialV == "" {
		s.Fail("no initial version specified")
	}
	s.upgradeParams.InitialVersion = initialV

	targetV := os.Getenv("TARGET_VERSION")
	if targetV == "" {
		s.upgradeParams.TargetVersion, err = s.upgradeManager.RetrieveUpgradeVersion()
		s.upgradeParams.TargetRepo = localRepository
		s.Require().NoError(err)
	} else {
		s.upgradeParams.TargetVersion = targetV
		s.upgradeParams.TargetRepo = localRepository
	}
	chainID := os.Getenv("CHAIN_ID")
	if chainID == "" {
		s.upgradeParams.ChainID = defaultChainID
	}
	skipFlag := os.Getenv("E2E_SKIP_CLEANUP")
	skipCleanup, err := strconv.ParseBool(skipFlag)
	s.Require().NoError(err, "invalid skip cleanup flag")
	s.upgradeParams.SkipCleanup = skipCleanup

	mountPath := os.Getenv("MOUNT_PATH")
	if mountPath == "" {
		s.Fail("no mount path specified")
	}
	s.upgradeParams.MountPath = mountPath
	s.T().Log("upgrade params: ", s.upgradeParams)
}
