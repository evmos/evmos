package types

import (
	"testing"

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
	newGen := NewGenesisState(DefaultParams(), uint64(0))

	testCases := []struct {
		name     string
		genState *GenesisState
		expPass  bool
	}{
		{
			"valid genesis constructor",
			&newGen,
			true,
		},
		{
			"default",
			DefaultGenesisState(),
			true,
		},
		{
			"valid genesis",
			&GenesisState{
				Params: DefaultParams(),
			},
			true,
		},
		{
			"valid genesis - with period",
			&GenesisState{
				Params: DefaultParams(),
				Period: uint64(0),
			},
			true,
		},
		{
			"empty genesis",
			&GenesisState{},
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
