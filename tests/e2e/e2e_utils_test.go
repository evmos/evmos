package e2e

import (
	"os"
	"strconv"
	"strings"
)

// versionSeparator is used to separate versions in the INITIAL_VERSION and TARGET_VERSION
// environment vars
const versionSeparator = "/"

// upgradeConfig defines a struct that contains the version and the source repository for an upgrade
type upgradeConfig struct {
	name    string
	version string
	repo    string
}

// loadUpgradeParams loads the parameters for the upgrade test suite from the environment
// variables
func (s *IntegrationTestSuite) loadUpgradeParams() {
	var err error
	// name defines the upgrade name to use in the proposal
	var name string
	// targetRepo is assigned to the remote repository by default and is changed to local if no
	// target version is given
	targetRepo := tharsisRepo
	// upgradesList contains the available upgrades in the app/upgrades folder
	var upgradesList []string
	// upgrades contains the slice of all upgrades that shall be executed
	var upgrades []upgradeConfig //nolint:prealloc
	// version defines the version to run the Evmos node with
	var version string
	// versions contains the slice of all versions that are run during the upgrade tests
	var versions []string

	initialV := os.Getenv("INITIAL_VERSION")
	if initialV == "" {
		upgradesList, err = s.upgradeManager.RetrieveUpgradesList()
		s.Require().NoError(err)
		// set the second-to-last upgrade as initial version
		versions = []string{upgradesList[len(upgradesList)-2]}
	} else {
		versions = strings.Split(initialV, versionSeparator)
	}

	// for all initial versions define the docker hub repo as the source
	for _, version := range versions {
		upgrades = append(upgrades, upgradeConfig{
			version,
			version,
			targetRepo,
		})
	}

	// Target version loading, if not specified manager gets the last one from app/upgrades folder
	// and sets target repository to local, otherwise tharsishq repo will be used
	targetV := os.Getenv("TARGET_VERSION")
	if targetV == "" {
		if upgradesList == nil {
			upgradesList, err = s.upgradeManager.RetrieveUpgradesList()
			s.Require().NoError(err)
		}
		name = upgradesList[len(upgradesList)-1]
		version = localVersionTag
	} else {
		name = targetV
		version = targetV
	}

	// Add the target version to the upgrades slice
	upgrades = append(upgrades, upgradeConfig{
		name,
		version,
		targetRepo,
	})
	s.upgradeParams.Upgrades = upgrades

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
