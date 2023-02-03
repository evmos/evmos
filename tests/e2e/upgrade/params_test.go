package upgrade

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// envVars is a helper struct to define the used environment variables
type envVars struct {
	initialVersion string
	targetVersion  string
	chainID        string
	skipCleanup    string
	mountPath      string
}

// TestLoadUpgradeParams tests the LoadUpgradeParams function
func TestLoadUpgradeParams(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err, "can't get current working directory")

	defaultMountPath := wd + "/build/:/root/"
	availableUpgrades, err := RetrieveUpgradesList(upgradesPath)
	require.NoError(t, err, "can't retrieve upgrades list")
	latestVersionName := availableUpgrades[len(availableUpgrades)-1]
	defaultInitialVersion := availableUpgrades[len(availableUpgrades)-2]

	testcases := []struct {
		name    string
		vars    envVars
		want    Params
		expPass bool
	}{
		{
			name: "pass - all params set - one initial version",
			vars: envVars{
				initialVersion: "v0.1.0",
				targetVersion:  "v0.2.0",
				chainID:        "evmos_9123-1",
				skipCleanup:    "true",
				mountPath:      "/tmp/evmos",
			},
			want: Params{
				MountPath: "/tmp/evmos",
				Versions: []VersionConfig{
					{"v0.1.0", "v0.1.0", tharsisRepo},
					{"v0.2.0", "v0.2.0", tharsisRepo},
				},
				ChainID:     "evmos_9123-1",
				WorkDirRoot: wd,
			},
			expPass: true,
		},
		{
			name: "pass - multiple initial versions, no target version",
			vars: envVars{
				initialVersion: "v0.1.0/v0.2.0",
			},
			want: Params{
				MountPath: defaultMountPath,
				Versions: []VersionConfig{
					{"v0.1.0", "v0.1.0", tharsisRepo},
					{"v0.2.0", "v0.2.0", tharsisRepo},
					{latestVersionName, LocalVersionTag, tharsisRepo},
				},
				ChainID:     defaultChainID,
				WorkDirRoot: wd,
			},
			expPass: true,
		},
		{
			name: "pass - no 'v' prefix in version string",
			vars: envVars{
				initialVersion: "0.1.0",
				targetVersion:  "0.2.0",
			},
			want: Params{
				MountPath: defaultMountPath,
				Versions: []VersionConfig{
					{"v0.1.0", "v0.1.0", tharsisRepo},
					{"v0.2.0", "v0.2.0", tharsisRepo},
				},
				ChainID:     defaultChainID,
				WorkDirRoot: wd,
			},
			expPass: true,
		},
		{
			name: "pass - release candidate version",
			vars: envVars{
				initialVersion: "v0.1.0-rc1",
				targetVersion:  "v0.2.0-rc2",
			},
			want: Params{
				MountPath: defaultMountPath,
				Versions: []VersionConfig{
					{"v0.1.0-rc1", "v0.1.0-rc1", tharsisRepo},
					{"v0.2.0-rc2", "v0.2.0-rc2", tharsisRepo},
				},
				ChainID:     defaultChainID,
				WorkDirRoot: wd,
			},
			expPass: true,
		},
		{
			name: "pass - no initial version",
			vars: envVars{},
			want: Params{
				MountPath: defaultMountPath,
				Versions: []VersionConfig{
					{defaultInitialVersion, defaultInitialVersion, tharsisRepo},
					{latestVersionName, LocalVersionTag, tharsisRepo},
				},
				ChainID:     defaultChainID,
				WorkDirRoot: wd,
			},
			expPass: true,
		},
		{
			name: "fail - separator in target version",
			vars: envVars{
				initialVersion: "v0.1.0",
				targetVersion:  "v0.2.0/v0.3.0",
			},
			want:    Params{},
			expPass: false,
		},
		{
			name: "fail - wrong version separator",
			vars: envVars{
				initialVersion: "v0.1.0|v0.2.0",
			},
			want:    Params{},
			expPass: false,
		},
		{
			name: "fail - invalid target version string",
			vars: envVars{
				targetVersion: "v@93bca",
			},
			want:    Params{},
			expPass: false,
		},
		{
			name: "fail - invalid initial version string",
			vars: envVars{
				initialVersion: "v@93bca",
			},
			want:    Params{},
			expPass: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("INITIAL_VERSION", tc.vars.initialVersion)
			t.Setenv("TARGET_VERSION", tc.vars.targetVersion)
			t.Setenv("CHAIN_ID", tc.vars.chainID)
			t.Setenv("SKIP_CLEANUP", tc.vars.skipCleanup)
			t.Setenv("MOUNT_PATH", tc.vars.mountPath)

			params, err := LoadUpgradeParams(upgradesPath)
			if tc.expPass {
				require.NoError(t, err, "LoadUpgradeParams() should not return an error")
			} else {
				require.Error(t, err, "LoadUpgradeParams() should return an error")
			}
			require.Equal(t, tc.want.ChainID, params.ChainID, "chain id differs from the expected value")
			require.Equal(t, tc.want.MountPath, params.MountPath, "mount path differs from the expected value")
			require.Equal(t, tc.want.Versions, params.Versions, "versions differ from the expected values")
			require.Equal(t, tc.want.WorkDirRoot, params.WorkDirRoot, "root working directory differs from the expected value")
			require.Equal(t, tc.want.SkipCleanup, params.SkipCleanup, "flag to skip cleanup differs from the expected value")
		})
	}
}
