package types

import (
	"testing"
	"time"

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
			NewGenesisState(NewParams(true, time.Hour)),
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
