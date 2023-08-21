package types_test

import (
	"testing"

	"github.com/evmos/evmos/v14/x/erc20/types"
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
	newGen := types.NewGenesisState(types.DefaultParams(), []types.TokenPair{})

	testCases := []struct {
		name     string
		genState *types.GenesisState
		expPass  bool
	}{
		{
			name:     "valid genesis constructor",
			genState: &newGen,
			expPass:  true,
		},
		{
			name:     "default",
			genState: types.DefaultGenesisState(),
			expPass:  true,
		},
		{
			name: "valid genesis",
			genState: &types.GenesisState{
				Params:     types.DefaultParams(),
				TokenPairs: []types.TokenPair{},
			},
			expPass: true,
		},
		{
			name: "valid genesis - with tokens pairs",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				TokenPairs: []types.TokenPair{
					{
						Erc20Address: "0xdac17f958d2ee523a2206206994597c13d831ec7",
						Denom:        "usdt",
						Enabled:      true,
					},
				},
			},
			expPass: true,
		},
		{
			name: "invalid genesis - duplicated token pair",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				TokenPairs: []types.TokenPair{
					{
						Erc20Address: "0xdac17f958d2ee523a2206206994597c13d831ec7",
						Denom:        "usdt",
						Enabled:      true,
					},
					{
						Erc20Address: "0xdac17f958d2ee523a2206206994597c13d831ec7",
						Denom:        "usdt",
						Enabled:      true,
					},
				},
			},
			expPass: false,
		},
		{
			name: "invalid genesis - duplicated token pair",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				TokenPairs: []types.TokenPair{
					{
						Erc20Address: "0xdac17f958d2ee523a2206206994597c13d831ec7",
						Denom:        "usdt",
						Enabled:      true,
					},
					{
						Erc20Address: "0xdac17f958d2ee523a2206206994597c13d831ec7",
						Denom:        "usdt2",
						Enabled:      true,
					},
				},
			},
			expPass: false,
		},
		{
			name: "invalid genesis - duplicated token pair",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				TokenPairs: []types.TokenPair{
					{
						Erc20Address: "0xdac17f958d2ee523a2206206994597c13d831ec7",
						Denom:        "usdt",
						Enabled:      true,
					},
					{
						Erc20Address: "0xB8c77482e45F1F44dE1745F52C74426C631bDD52",
						Denom:        "usdt",
						Enabled:      true,
					},
				},
			},
			expPass: false,
		},
		{
			name: "invalid genesis - invalid token pair",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				TokenPairs: []types.TokenPair{
					{
						Erc20Address: "0xinvalidaddress",
						Denom:        "bad",
						Enabled:      true,
					},
				},
			},
			expPass: false,
		},
		{
			// Voting period cant be zero
			name:     "empty genesis",
			genState: &types.GenesisState{},
			expPass:  true,
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
