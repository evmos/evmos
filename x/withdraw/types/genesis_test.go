package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesisValidate(t *testing.T) {
	testCases := []struct {
		name     string
		genesis  GenesisState
		expError bool
	}{
		{
			"empty genesis",
			GenesisState{},
			false,
		},
		{
			"default genesis",
			*DefaultGenesisState(),
			false,
		},
		{
			"custom genesis",
			NewGenesisState(NewParams(true)),
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.genesis.Validate()
		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}
