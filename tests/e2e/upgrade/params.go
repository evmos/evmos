// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package upgrade

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Params defines the parameters for the upgrade test suite
type Params struct {
	// MountPath defines the path where the docker container is mounted
	MountPath string
	// Versions defines the slice of versions that are run during the upgrade tests
	Versions []VersionConfig
	// ChainID defines the chain ID used for the upgrade tests
	ChainID string
	// SkipCleanup defines if the docker containers are removed after the tests
	SkipCleanup bool
	// WorkDirRoot defines the working directory
	WorkDirRoot string
}

// VersionConfig defines a struct that contains the version and the source repository for an upgrade
type VersionConfig struct {
	// UpgradeName defines the upgrade name to use in the proposal
	UpgradeName string
	// ImageTag defines the version tag to use in the docker image
	ImageTag string
	// ImageName defines the image name for the docker image
	ImageName string
}

// LoadUpgradeParams loads the upgrade parameters from the environment variables
func LoadUpgradeParams(upgradesFolder string) (Params, error) {
	var (
		// allowedVersionPattern defines the regex pattern for a semver version (including release candidates)
		allowedVersionPattern = `v*\d+\.\d\.\d(-rc\d+)*`
		// allowedVersionSinglePattern defines the regex pattern for a single version
		allowedVersionSinglePattern = fmt.Sprintf(`^%s$`, allowedVersionPattern)
		// allowedVersionListPattern defines the regex pattern for a list of versions
		allowedVersionListPattern = fmt.Sprintf(`^(%s*%s)+$`, versionSeparator, allowedVersionPattern)
		// err is the captured error variable
		err error
		// upgradeName defines the upgrade name to use in the proposal
		upgradeName string
		// upgradesList contains the available upgrades in the app/upgrades folder
		upgradesList []string
		// versionTag is a string to store the processed version tags (e.g. v10.0.1)
		versionTag string
		// versionTags contains the slice of all version tags that are run during the upgrade tests
		versionTags []string
	)

	initialV := os.Getenv("INITIAL_VERSION")
	if initialV == "" {
		upgradesList, err = RetrieveUpgradesList(upgradesFolder)
		if err != nil {
			return Params{}, fmt.Errorf("failed to retrieve the list of upgrades: %w", err)
		}
		// set the second-to-last upgrade as initial version
		versionTags = []string{upgradesList[len(upgradesList)-2]}
	} else {
		if !regexp.MustCompile(allowedVersionListPattern).MatchString(initialV) {
			return Params{}, fmt.Errorf("invalid initial version: %s", initialV)
		}
		versionTags = strings.Split(initialV, versionSeparator)
	}

	// versions contains the slice of all versions that shall be executed
	versions := make([]VersionConfig, 0, len(versionTags))

	// for all initial versions the docker hub image is used
	for _, versionTag = range versionTags {
		if !strings.Contains(versionTag, "v") {
			versionTag = "v" + versionTag
		}
		versions = append(versions, VersionConfig{
			UpgradeName: versionTag,
			ImageTag:    versionTag,
			ImageName:   tharsisRepo,
		})
	}

	// When a target version is specified, it is used and the tharsishq DockerHub repo used.
	// If no target version is specified, the last upgrade in the app/upgrades folder is used
	// and a name for the local image is assigned.
	targetV := os.Getenv("TARGET_VERSION")
	if targetV == "" {
		if upgradesList == nil {
			upgradesList, err = RetrieveUpgradesList(upgradesFolder)
			if err != nil {
				return Params{}, fmt.Errorf("failed to retrieve the list of upgrades: %w", err)
			}
		}
		upgradeName = upgradesList[len(upgradesList)-1]
		versionTag = LocalVersionTag
	} else {
		if !regexp.MustCompile(allowedVersionSinglePattern).MatchString(targetV) {
			return Params{}, fmt.Errorf("invalid target version: %s", targetV)
		}
		if !strings.Contains(targetV, "v") {
			targetV = "v" + targetV
		}
		upgradeName = targetV
		versionTag = targetV
	}

	// Add the target version to the versions slice
	versions = append(versions, VersionConfig{
		upgradeName,
		versionTag,
		tharsisRepo,
	})

	// If chain ID is not specified, the default value from the constants file will be used in upgrade-init.sh
	chainID := os.Getenv("CHAIN_ID")
	if chainID == "" {
		chainID = defaultChainID
	}

	skipFlag := os.Getenv("E2E_SKIP_CLEANUP")
	skipCleanup, err := strconv.ParseBool(skipFlag)
	if err != nil {
		skipCleanup = false
	}

	workDir, err := os.Getwd()
	if err != nil {
		return Params{}, err
	}
	workDir = strings.TrimSuffix(workDir, "/tests/e2e")

	mountPath := os.Getenv("MOUNT_PATH")
	if mountPath == "" {
		mountPath = workDir + "/build/:/root/"
	}

	params := Params{
		MountPath:   mountPath,
		Versions:    versions,
		ChainID:     chainID,
		SkipCleanup: skipCleanup,
		WorkDirRoot: workDir,
	}

	return params, nil
}
