// This file contains unit tests for the e2e package.
package upgrade

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestCheckLegacyProposal tests the checkLegacyProposal function with different version strings
func TestCheckLegacyProposal(t *testing.T) {
	var legacyProposal bool

	testCases := []struct {
		Name string
		Ver  string
		Exp  bool
	}{
		{
			Name: "legacy proposal - v10.0.1",
			Ver:  "v10.0.1",
			Exp:  true,
		},
		{
			Name: "normal proposal - v9.1.0",
			Ver:  "v9.1.0",
			Exp:  false,
		},
		{
			Name: "normal proposal - version with whitespace - v9.1.0",
			Ver:  "\tv9.1.0 ",
			Exp:  false,
		},
		{
			Name: "normal proposal - version without v - 9.1.0",
			Ver:  "9.1.0",
			Exp:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			legacyProposal = CheckLegacyProposal(tc.Ver)
			require.Equal(t, legacyProposal, tc.Exp, "expected: %v, got: %v", tc.Exp, legacyProposal)
		})
	}
}

// TestEvmosVersionsLess tests the EvmosVersions type's Less method with
// different version strings
func TestEvmosVersionsLess(t *testing.T) {
	var version EvmosVersions

	testCases := []struct {
		Name string
		Ver  string
		Exp  bool
	}{
		{
			Name: "higher - v10.0.1",
			Ver:  "v10.0.1",
			Exp:  false,
		},
		{
			Name: "lower - v9.1.0",
			Ver:  "v9.1.0",
			Exp:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			version = []string{tc.Ver, "v10.0.0"}
			require.Equal(t, version.Less(0, 1), tc.Exp, "expected: %v, got: %v", tc.Exp, version)
		})
	}
}

// TestEvmosVersionsSwap tests the EvmosVersions type's Swap method
func TestEvmosVersionsSwap(t *testing.T) {
	var version EvmosVersions
	value := "v9.1.0"
	version = []string{value, "v10.0.0"}
	version.Swap(0, 1)
	require.Equal(t, value, version[1], "expected: %v, got: %v", value, version[1])
}

// TestEvmosVersionsLen tests the EvmosVersions type's Len method
func TestEvmosVersionsLen(t *testing.T) {
	var version EvmosVersions = []string{"v9.1.0", "v10.0.0"}
	require.Equal(t, 2, version.Len(), "expected: %v, got: %v", 2, version.Len())
}

// TestRetrieveUpgradesList tests if the list of available upgrades in the codebase
// can be correctly retrieved
func TestRetrieveUpgradesList(t *testing.T) {
	upgradeList, err := RetrieveUpgradesList("../../../app/upgrades")
	require.NoError(t, err, "expected no error while retrieving upgrade list")
	require.NotEmpty(t, upgradeList, "expected upgrade list to be non-empty")

	// check if all entries in the list match a semantic versioning pattern
	for _, upgrade := range upgradeList {
		require.Regexp(t, `^v\d+\.\d+\.\d+(-rc\d+)*$`, upgrade, "expected upgrade version to be in semantic versioning format")
	}
}
