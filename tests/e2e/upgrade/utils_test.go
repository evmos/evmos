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

// TestByVersion tests the EvmosVersions type with different version strings
func TestByVersion(t *testing.T) {
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
