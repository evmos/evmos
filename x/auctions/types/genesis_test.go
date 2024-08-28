package types_test

import (
	"testing"

	"cosmossdk.io/math"

	"github.com/evmos/evmos/v19/x/auctions/types"
	"github.com/stretchr/testify/assert"
)

func TestGenesisStateValidate(t *testing.T) {
	testCases := []struct {
		name   string
		mutate func(*types.GenesisState)
		expErr bool
	}{
		{
			"valid default",
			func(*types.GenesisState) {},
			false,
		},
		{
			"fail - wrong bid denom",
			func(genesis *types.GenesisState) { genesis.Bid.BidValue.Denom = "uatom" },
			true,
		},
		{
			"fail - negative amount",
			func(genesis *types.GenesisState) { genesis.Bid.BidValue.Amount = math.NewInt(-999) },
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			genesisState := types.DefaultGenesisState()

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
