package keeper_test

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/contracts"
	"github.com/evmos/evmos/v19/testutil"
	utiltx "github.com/evmos/evmos/v19/testutil/tx"
	"github.com/evmos/evmos/v19/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
)

const erc20Decimals = uint8(18)

// DeployContract deploys the ERC20MinterBurnerDecimalsContract.
func (suite *KeeperTestSuite) DeployContract(name, symbol string, decimals uint8) (common.Address, error) {
	suite.Commit()
	addr, err := testutil.DeployContract(
		suite.ctx,
		suite.app,
		suite.priv,
		suite.queryClient,
		contracts.ERC20MinterBurnerDecimalsContract,
		name, symbol, decimals,
	)
	suite.Commit()
	return addr, err
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
		account := utiltx.GenerateAddress()

		res, err := suite.app.EvmKeeper.CallEVM(suite.ctx, erc20, types.ModuleAddress, contract, true, tc.method, account)
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
				account := utiltx.GenerateAddress()
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
				account := utiltx.GenerateAddress()
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
				data := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...) //nolint:gocritic
				return data, nil
			},
			true,
		},
		{
			"fail deploy",
			types.ModuleAddress,
			func() ([]byte, *common.Address) {
				params := suite.app.EvmKeeper.GetParams(suite.ctx)
				params.AccessControl.Create = evmtypes.AccessControlType{
					AccessType: evmtypes.AccessTypeRestricted,
				}
				_ = suite.app.EvmKeeper.SetParams(suite.ctx, params)
				ctorArgs, _ := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("", "test", "test", uint8(18))
				data := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...) //nolint:gocritic
				return data, nil
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			data, contract := tc.malleate()

			res, err := suite.app.EvmKeeper.CallEVMWithData(suite.ctx, tc.from, contract, data, true)
			if tc.expPass {
				suite.Require().IsTypef(&evmtypes.MsgEthereumTxResponse{}, res, tc.name)
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
