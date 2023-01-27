package e2e

import (
	"os"
	"strconv"
	"strings"

	"github.com/evmos/evmos/v11/tests/e2e/upgrade"
)

// upgradesPath is the relative path from this file to the app/upgrades folder
const upgradesPath = "../../app/upgrades"

// versionSeparator is used to separate versions in the INITIAL_VERSION and TARGET_VERSION
// environment vars
const versionSeparator = "/"

// versionConfig defines a struct that contains the version and the source repository for an upgrade
type versionConfig struct {
	name string
	tag  string
	repo string
}

// loadUpgradeParams loads the parameters for the upgrade test suite from the environment
// variables
func (s *IntegrationTestSuite) loadUpgradeParams() {
	var (
		err error
		// name defines the upgrade name to use in the proposal
		name string
		// targetRepo is assigned to the remote repository by default and is changed to local if no
		// target version is given
		targetRepo = tharsisRepo
		// upgradesList contains the available upgrades in the app/upgrades folder
		upgradesList []string
		// versionTag is a string to store the processed version tags (e.g. v10.0.1)
		versionTag string
		// versionTags contains the slice of all version tags that are run during the upgrade tests
		versionTags []string
	)

	initialV := os.Getenv("INITIAL_VERSION")
	if initialV == "" {
		upgradesList, err = upgrade.RetrieveUpgradesList(upgradesPath)
		s.Require().NoError(err)
		// set the second-to-last upgrade as initial version
		versionTags = []string{upgradesList[len(upgradesList)-2]}
	} else {
		versionTags = strings.Split(initialV, versionSeparator)
	}

	// versions contains the slice of all versions that shall be executed
	versions := make([]versionConfig, 0, len(versionTags))

	// for all initial versions define the docker hub repo as the source
	for _, versionTag = range versionTags {
		versions = append(versions, versionConfig{
			name: versionTag,
			tag:  versionTag,
			repo: targetRepo,
		})
	}

	// Target version loading, if not specified manager gets the last one from app/upgrades folder
	// and sets target repository to local, otherwise tharsishq repo will be used
	targetV := os.Getenv("TARGET_VERSION")
	if targetV == "" {
		if upgradesList == nil {
			upgradesList, err = upgrade.RetrieveUpgradesList(upgradesPath)
			s.Require().NoError(err)
		}
		name = upgradesList[len(upgradesList)-1]
		versionTag = localVersionTag
	} else {
		name = targetV
		versionTag = targetV
	}

	// Add the target version to the versions slice
	versions = append(versions, versionConfig{
		name,
		versionTag,
		targetRepo,
	})
	s.upgradeParams.Versions = versions

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
