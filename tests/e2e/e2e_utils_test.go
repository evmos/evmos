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

	// Target version loading, if not specified manager gets the last one from app/upgrades folder
	// and sets target repository to local, otherwise tharsishq repo will be used
	targetV := os.Getenv("TARGET_VERSION")
	if targetV == "" {
		s.upgradeParams.SoftwareUpgradeVersion, err = s.upgradeManager.RetrieveUpgradeVersion()
		s.Require().NoError(err)
		s.upgradeParams.TargetVersion = localVersionTag
		s.upgradeParams.TargetRepo = localRepository
	} else {
		s.upgradeParams.TargetVersion = targetV
		s.upgradeParams.SoftwareUpgradeVersion = targetV
		s.upgradeParams.TargetRepo = tharsisRepo
	}

	// If chain ID is not specified, 'evmos_9000-1' will be used in upgrade-init.sh
	chID := os.Getenv("CHAIN_ID")
	if chID == "" {
		chID = defaultChainID
	}
	s.upgradeParams.ChainID = chID

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
