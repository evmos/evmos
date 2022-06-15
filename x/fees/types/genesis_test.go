package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"github.com/tharsis/ethermint/tests"
)

type GenesisTestSuite struct {
	suite.Suite
	address1 string
	address2 string
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) SetupTest() {
	suite.address1 = sdk.AccAddress(tests.GenerateAddress().Bytes()).String()
	suite.address2 = sdk.AccAddress(tests.GenerateAddress().Bytes()).String()
}

func (suite *GenesisTestSuite) TestValidateGenesis() {
	newGen := NewGenesisState(DefaultParams(), []Fee{})
	testCases := []struct {
		name     string
		genState *GenesisState
		expPass  bool
	}{
		{
			name:     "valid genesis constructor",
			genState: &newGen,
			expPass:  true,
		},
		{
			name:     "default",
			genState: DefaultGenesisState(),
			expPass:  true,
		},
		{
			name: "valid genesis",
			genState: &GenesisState{
				Params: DefaultParams(),
				Fees:   []Fee{},
			},
			expPass: true,
		},
		{
			name: "valid genesis - with fee information",
			genState: &GenesisState{
				Params: DefaultParams(),
				Fees: []Fee{
					{
						ContractAddress: "0xdac17f958d2ee523a2206206994597c13d831ec7",
						DeployerAddress: suite.address1,
					},
					{
						ContractAddress: "0xdac17f958d2ee523a2206206994597c13d831ec8",
						DeployerAddress: suite.address2,
						WithdrawAddress: suite.address2,
					},
				},
			},
			expPass: true,
		},
		{
			name:     "empty genesis",
			genState: &GenesisState{},
			expPass:  false,
		},
		{
			name: "invalid genesis - duplicated fee info",
			genState: &GenesisState{
				Params: DefaultParams(),
				Fees: []Fee{
					{
						ContractAddress: "0xdac17f958d2ee523a2206206994597c13d831ec7",
						DeployerAddress: suite.address1,
					},
					{
						ContractAddress: "0xdac17f958d2ee523a2206206994597c13d831ec7",
						DeployerAddress: suite.address1,
					},
				},
			},
			expPass: false,
		},
		{
			name: "invalid genesis - duplicated fee info 2",
			genState: &GenesisState{
				Params: DefaultParams(),
				Fees: []Fee{
					{
						ContractAddress: "0xdac17f958d2ee523a2206206994597c13d831ec7",
						DeployerAddress: suite.address1,
					},
					{
						ContractAddress: "0xdac17f958d2ee523a2206206994597c13d831ec7",
						DeployerAddress: suite.address2,
					},
				},
			},
			expPass: false,
		},
		{
			name: "invalid genesis - invalid contract address",
			genState: &GenesisState{
				Params: DefaultParams(),
				Fees: []Fee{
					{
						ContractAddress: suite.address1,
						DeployerAddress: suite.address1,
					},
				},
			},
			expPass: false,
		},
		{
			name: "invalid genesis - invalid deployer address",
			genState: &GenesisState{
				Params: DefaultParams(),
				Fees: []Fee{
					{
						ContractAddress: "0xdac17f958d2ee523a2206206994597c13d831ec7",
						DeployerAddress: "0xdac17f958d2ee523a2206206994597c13d831ec7",
					},
				},
			},
			expPass: false,
		},
		{
			name: "invalid genesis - invalid withdraw address",
			genState: &GenesisState{
				Params: DefaultParams(),
				Fees: []Fee{
					{
						ContractAddress: "0xdac17f958d2ee523a2206206994597c13d831ec7",
						DeployerAddress: suite.address1,
						WithdrawAddress: "0xdac17f958d2ee523a2206206994597c13d831ec7",
					},
				},
			},
			expPass: false,
		},
		{
			name: "invalid genesis - invalid params",
			genState: &GenesisState{
				Params: DefaultParams(),
				Fees: []Fee{
					{
						ContractAddress: "0xdac17f958d2ee523a2206206994597c13d831ec7",
						DeployerAddress: suite.address1,
						WithdrawAddress: "withdraw",
					},
				},
			},
			expPass: false,
		},
	}

	for _, tc := range testCases {
		err := tc.genState.Validate()
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}
