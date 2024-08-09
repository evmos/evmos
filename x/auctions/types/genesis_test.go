package types

import (
	"testing"

	"cosmossdk.io/math"

	"github.com/stretchr/testify/assert"
)

func TestGenesisStateValidate(t *testing.T) {
	testCases := []struct {
		name   string
		mutate func(*GenesisState)
		expErr bool
	}{
		{
			"valid default",
			func(*GenesisState) {},
			false,
		},
		{
			"invalid wrong bid denom",
			func(genesis *GenesisState) { genesis.Bid.Amount.Denom = "uatom" },
			true,
		},
		{
			"invalid negative amount",
			func(genesis *GenesisState) { genesis.Bid.Amount.Amount = math.NewInt(-999) },
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			genesisState := DefaultGenesisState()

			tc.mutate(genesisState)

			err := genesisState.Validate()

			if tc.expErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
