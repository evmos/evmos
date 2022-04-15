package keeper_test

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/ethermint/tests"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/tharsis/evmos/v3/contracts"
	"github.com/tharsis/evmos/v3/x/erc20/types"
)

func (suite *KeeperTestSuite) TestQueryERC20() {
	var contract common.Address
	testCases := []struct {
		name     string
		malleate func()
		res      bool
	}{
		{
			"erc20 not deployed",
			func() { contract = common.Address{} },
			false,
		},
		{
			"ok",
			func() { contract, _ = suite.DeployContract("coin", "token", erc20Decimals) },
			true,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset

		tc.malleate()

		res, err := suite.app.Erc20Keeper.QueryERC20(suite.ctx, contract)
		if tc.res {
			suite.Require().NoError(err)
			suite.Require().Equal(
				types.ERC20Data{Name: "coin", Symbol: "token", Decimals: erc20Decimals},
				res,
			)
		} else {
			suite.Require().Error(err)
		}
	}
}

func (suite *KeeperTestSuite) TestCallEVM() {
	testCases := []struct {
		name    string
		method  string
		expPass bool
	}{
		{
			"unknown method",
			"",
			false,
		},
		{
			"pass",
			"balanceOf",
			true,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset

		erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
		contract, err := suite.DeployContract("coin", "token", erc20Decimals)
		suite.Require().NoError(err)
		account := tests.GenerateAddress()

		res, err := suite.app.Erc20Keeper.CallEVM(suite.ctx, erc20, types.ModuleAddress, contract, true, tc.method, account)
		if tc.expPass {
			suite.Require().IsTypef(&evmtypes.MsgEthereumTxResponse{}, res, tc.name)
			suite.Require().NoError(err)
		} else {
			suite.Require().Error(err)
		}
	}
}

func (suite *KeeperTestSuite) TestCallEVMWithData() {
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
	testCases := []struct {
		name     string
		from     common.Address
		malleate func() ([]byte, *common.Address)
		expPass  bool
	}{
		{
			"unknown method",
			types.ModuleAddress,
			func() ([]byte, *common.Address) {
				contract, err := suite.DeployContract("coin", "token", erc20Decimals)
				suite.Require().NoError(err)
				account := tests.GenerateAddress()
				data, _ := erc20.Pack("", account)
				return data, &contract
			},
			false,
		},
		{
			"pass",
			types.ModuleAddress,
			func() ([]byte, *common.Address) {
				contract, err := suite.DeployContract("coin", "token", erc20Decimals)
				suite.Require().NoError(err)
				account := tests.GenerateAddress()
				data, _ := erc20.Pack("balanceOf", account)
				return data, &contract
			},
			true,
		},
		{
			"fail empty data",
			types.ModuleAddress,
			func() ([]byte, *common.Address) {
				contract, err := suite.DeployContract("coin", "token", erc20Decimals)
				suite.Require().NoError(err)
				return []byte{}, &contract
			},
			false,
		},

		{
			"fail empty sender",
			common.Address{},
			func() ([]byte, *common.Address) {
				contract, err := suite.DeployContract("coin", "token", erc20Decimals)
				suite.Require().NoError(err)
				return []byte{}, &contract
			},
			false,
		},
		{
			"deploy",
			types.ModuleAddress,
			func() ([]byte, *common.Address) {
				ctorArgs, _ := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("", "test", "test", uint8(18))
				data := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...)
				return data, nil
			},
			true,
		},
		{
			"fail deploy",
			types.ModuleAddress,
			func() ([]byte, *common.Address) {
				params := suite.app.EvmKeeper.GetParams(suite.ctx)
				params.EnableCreate = false
				suite.app.EvmKeeper.SetParams(suite.ctx, params)
				ctorArgs, _ := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("", "test", "test", uint8(18))
				data := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...)
				return data, nil
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			data, contract := tc.malleate()

			res, err := suite.app.Erc20Keeper.CallEVMWithData(suite.ctx, tc.from, contract, data, true)
			if tc.expPass {
				suite.Require().IsTypef(&evmtypes.MsgEthereumTxResponse{}, res, tc.name)
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
