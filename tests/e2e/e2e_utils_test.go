package e2e

import (
	"os"
	"strconv"
)

func (s *IntegrationTestSuite) loadUpgradeParams() {
	var err error

	preV := os.Getenv("INITIAL_VERSION")
	if preV == "" {
		s.Fail("no initial version specified")
	}
	s.upgradeParams.InitialVersion = preV

	postV := os.Getenv("TARGET_VERSION")
	if postV == "" {
		s.upgradeParams.InitialVersion, err = s.upgradeManager.RetrieveUpgradeVersion()
		s.Require().NoError(err)
	}
	s.upgradeParams.TargetVersion = postV

	skipFlag := os.Getenv("E2E_SKIP_CLEANUP")
	skipCleanup, err := strconv.ParseBool(skipFlag)
	s.Require().NoError(err, "invalid skip cleanup flag")
	s.upgradeParams.SkipCleanup = skipCleanup

	mountPath := os.Getenv("MOUNT_PATH")
	if postV == "" {
		s.Fail("no mount path specified")
	}
	s.upgradeParams.MountPath = mountPath
	s.T().Log("upgrade params: ", s.upgradeParams)
}
