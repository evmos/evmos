package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	newGen := NewGenesisState(DefaultParams(), []Incentive{})

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
				Params:     DefaultParams(),
				Incentives: []Incentive{},
			},
			true,
		},
		// TODO: Dec coin amount is supposed to be in percent but only accepts int
		{
			"valid genesis - with incentives",
			&GenesisState{
				Params: DefaultParams(),
				Incentives: []Incentive{
					{
						Contract:    "0xdac17f958d2ee523a2206206994597c13d831ec7",
						Allocations: sdk.DecCoins{sdk.NewDecCoin("aphoton", sdk.NewInt(1))},
						Epochs:      10,
						StartTime:   time.Now(),
					},
				},
			},
			true,
		},
		{
			"invalid genesis - duplicated incentive",
			&GenesisState{
				Params: DefaultParams(),
				Incentives: []Incentive{
					{
						Contract:    "0xdac17f958d2ee523a2206206994597c13d831ec7",
						Allocations: sdk.DecCoins{sdk.NewDecCoin("aphoton", sdk.NewInt(1))},
						Epochs:      10,
						StartTime:   time.Now(),
					},
					{
						Contract:    "0xdac17f958d2ee523a2206206994597c13d831ec7",
						Allocations: sdk.DecCoins{sdk.NewDecCoin("aphoton", sdk.NewInt(1))},
						Epochs:      10,
						StartTime:   time.Now(),
					},
				},
			},
			false,
		},
		{
			"invalid genesis - invalid incentive",
			&GenesisState{
				Params: DefaultParams(),
				Incentives: []Incentive{
					{
						Contract:    "0xinvalidaddress",
						Allocations: sdk.DecCoins{sdk.NewDecCoin("aphoton", sdk.NewInt(1))},
						Epochs:      10,
						StartTime:   time.Now(),
					},
				},
			},
			false,
		},
		{
			// Voting period cant be zero
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
