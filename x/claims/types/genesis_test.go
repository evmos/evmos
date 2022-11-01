package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evoblockchain/ethermint/tests"
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
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())

	testCases := []struct {
		name     string
		genState *GenesisState
		expPass  bool
	}{
		{
			name:     "default",
			genState: DefaultGenesis(),
			expPass:  true,
		},
		{
			name: "valid genesis",
			genState: &GenesisState{
				Params:        DefaultParams(),
				ClaimsRecords: []ClaimsRecordAddress{},
			},
			expPass: true,
		},
		{
			name: "valid genesis - with claim record",
			genState: &GenesisState{
				Params: DefaultParams(),
				ClaimsRecords: []ClaimsRecordAddress{
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
			genState: &GenesisState{
				Params: DefaultParams(),
				ClaimsRecords: []ClaimsRecordAddress{
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
			genState: &GenesisState{
				Params: DefaultParams(),
				ClaimsRecords: []ClaimsRecordAddress{
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
			genState: &GenesisState{
				Params: DefaultParams(),
				ClaimsRecords: []ClaimsRecordAddress{
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
			genState: &GenesisState{},
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
