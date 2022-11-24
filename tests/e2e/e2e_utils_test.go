package e2e

import (
	"os"
	"strconv"
	"strings"
)

func (s *IntegrationTestSuite) loadUpgradeParams() {
	var err error

	initialV := os.Getenv("INITIAL_VERSION")
	if initialV == "" {
		upgradesList, err := s.upgradeManager.RetrieveUpgradesList()
		s.Require().NoError(err)
		// set the pre-last upgrade is upgrade list
		s.upgradeParams.InitialVersion = upgradesList[len(upgradesList)-2]
	} else {
		s.upgradeParams.InitialVersion = initialV
	}

	// Target version loading, if not specified manager gets the last one from app/upgrades folder
	// and sets target repository to local, otherwise tharsishq repo will be used
	targetV := os.Getenv("TARGET_VERSION")
	if targetV == "" {
		upgradesList, err := s.upgradeManager.RetrieveUpgradesList()
		s.Require().NoError(err)
		// set the last upgrade is upgrade list
		s.upgradeParams.SoftwareUpgradeVersion = upgradesList[len(upgradesList)-1]
		s.upgradeParams.TargetVersion = localVersionTag
		s.upgradeParams.TargetRepo = tharsisRepo
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
	if err != nil {
		skipCleanup = false
	}
	s.upgradeParams.SkipCleanup = skipCleanup

	wd, err := os.Getwd()
	s.Require().NoError(err)
	s.upgradeParams.WDRoot = strings.TrimSuffix(wd, "/tests/e2e")

	mountPath := os.Getenv("MOUNT_PATH")
	if mountPath == "" {

		mountPath = s.upgradeParams.WDRoot + "/build/:/root/"
	}
	s.upgradeParams.MountPath = mountPath
	s.T().Logf("upgrade params: %+v\n", s.upgradeParams)
}
