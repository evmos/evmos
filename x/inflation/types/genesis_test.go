package types

import (
	"testing"

	epochstypes "github.com/evmos/evmos/v14/x/epochs/types"
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
	// Team Address needs to be set manually at Genesis
	validParams := DefaultParams()

	newGen := NewGenesisState(validParams, uint64(0), epochstypes.DayEpochID, 365, 0)

	testCases := []struct {
		name     string
		genState *GenesisState
		expPass  bool
	}{
		{
			"empty genesis",
			&GenesisState{},
			false,
		},
		{
			"invalid default genesis",
			DefaultGenesisState(),
			true,
		},
		{
			"valid genesis constructor",
			&newGen,
			true,
		},
		{
			"valid genesis",
			&GenesisState{
				Params:          validParams,
				Period:          uint64(5),
				EpochIdentifier: epochstypes.DayEpochID,
				EpochsPerPeriod: 365,
			},
			true,
		},
		{
			"invalid genesis",
			&GenesisState{
				Params: validParams,
			},
			false,
		},
		{
			"invalid genesis - empty eporchIdentifier",
			&GenesisState{
				Params:          validParams,
				Period:          uint64(5),
				EpochIdentifier: "",
				EpochsPerPeriod: 365,
			},
			false,
		},
		{
			"invalid genesis - zero epochsPerPerid",
			&GenesisState{
				Params:          validParams,
				Period:          uint64(5),
				EpochIdentifier: epochstypes.DayEpochID,
				EpochsPerPeriod: 0,
			},
			false,
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
