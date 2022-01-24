package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tharsis/ethermint/tests"
)

func TestGenesisStateValidate(t *testing.T) {
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes()).String()

	testCases := []struct {
		name     string
		gs       *GenesisState
		expError bool
	}{
		{
			"valid - default params",
			DefaultGenesis(),
			false,
		},
		{
			"invalid - empty literal",
			&GenesisState{},
			true,
		},
		{
			"invalid claim record",
			&GenesisState{
				Params: DefaultParams(),
				ClaimRecords: []ClaimRecordAddress{
					{},
				},
			},
			true,
		},
		{
			"duplicated claim records",
			&GenesisState{
				Params: DefaultParams(),
				ClaimRecords: []ClaimRecordAddress{
					{
						Address:                addr,
						InitialClaimableAmount: sdk.NewInt(100),
						ActionsCompleted:       []bool{false, false, false, false},
					},
					{
						Address:                addr,
						InitialClaimableAmount: sdk.NewInt(10),
						ActionsCompleted:       []bool{true, true, true, true},
					},
				},
			},
			true,
		},
		{
			"invalid params",
			&GenesisState{
				Params: Params{},
			},
			true,
		},
	}

	for _, tc := range testCases {
		err := tc.gs.Validate()
		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}
