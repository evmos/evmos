package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	utiltx "github.com/evmos/evmos/v11/testutil/tx"
	"github.com/evmos/evmos/v11/x/claims/types"

	"github.com/stretchr/testify/suite"
)

type GenesisTestSuite struct {
	suite.Suite
}

func (suite *GenesisTestSuite) SetupTest() {
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestValidateGenesis() {
	addr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())

	testCases := []struct {
		name     string
		genState *types.GenesisState
		expPass  bool
	}{
		{
			name:     "default",
			genState: types.DefaultGenesis(),
			expPass:  true,
		},
		{
			name: "valid genesis",
			genState: &types.GenesisState{
				Params:        types.DefaultParams(),
				ClaimsRecords: []types.ClaimsRecordAddress{},
			},
			expPass: true,
		},
		{
			name: "valid genesis - with claim record",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				ClaimsRecords: []types.ClaimsRecordAddress{
					{
						Address:                addr.String(),
						InitialClaimableAmount: sdk.NewInt(1),
						ActionsCompleted:       []bool{true, true, false, false},
					},
				},
			},
			expPass: true,
		},
		{
			name: "invalid genesis - duplicated claim record",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				ClaimsRecords: []types.ClaimsRecordAddress{
					{
						Address:                addr.String(),
						InitialClaimableAmount: sdk.NewInt(1),
						ActionsCompleted:       []bool{true, true, false, false},
					},
					{
						Address:                addr.String(),
						InitialClaimableAmount: sdk.NewInt(1),
						ActionsCompleted:       []bool{true, true, false, false},
					},
				},
			},
			expPass: false,
		},

		{
			name: "invalid genesis - invalid address",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				ClaimsRecords: []types.ClaimsRecordAddress{
					{
						Address:                "badaddress",
						InitialClaimableAmount: sdk.NewInt(1),
						ActionsCompleted:       []bool{true, true, false, false},
					},
				},
			},
			expPass: false,
		},
		{
			name: "invalid genesis - invalid claimable amount",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				ClaimsRecords: []types.ClaimsRecordAddress{
					{
						Address:                addr.String(),
						InitialClaimableAmount: sdk.NewInt(-100),
						ActionsCompleted:       []bool{true, true, false, false},
					},
				},
			},
			expPass: false,
		},
		{
			// duration of decay must be positive
			name:     "empty genesis",
			genState: &types.GenesisState{},
			expPass:  false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		err := tc.genState.Validate()
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}
