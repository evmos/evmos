package keeper_test

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
<<<<<<< HEAD
	"github.com/evmos/evmos/v19/contracts"
	utiltx "github.com/evmos/evmos/v19/testutil/tx"
	"github.com/evmos/evmos/v19/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
)

func (suite *KeeperTestSuite) TestCallEVM() {
	wevmosContract := common.HexToAddress(types.WEVMOSContractMainnet)
=======
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
>>>>>>> main
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
<<<<<<< HEAD
		account := utiltx.GenerateAddress()
		res, err := suite.network.App.EvmKeeper.CallEVM(suite.network.GetContext(), erc20, types.ModuleAddress, wevmosContract, false, tc.method, account)
=======
		contract, err := suite.DeployContract("coin", "token", erc20Decimals)
		suite.Require().NoError(err)
		account := utiltx.GenerateAddress()

		res, err := suite.app.EvmKeeper.CallEVM(suite.ctx, erc20, types.ModuleAddress, contract, true, tc.method, account)
>>>>>>> main
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
<<<<<<< HEAD
	wevmosContract := common.HexToAddress(types.WEVMOSContractMainnet)
	testCases := []struct {
		name     string
		from     common.Address
		malleate func() []byte
		deploy   bool
=======
	testCases := []struct {
		name     string
		from     common.Address
		malleate func() ([]byte, *common.Address)
>>>>>>> main
		expPass  bool
	}{
		{
			"unknown method",
			types.ModuleAddress,
<<<<<<< HEAD
			func() []byte {
				account := utiltx.GenerateAddress()
				data, _ := erc20.Pack("", account)
				return data
			},
			false,
			false,
=======
			func() ([]byte, *common.Address) {
				contract, err := suite.DeployContract("coin", "token", erc20Decimals)
				suite.Require().NoError(err)
				account := utiltx.GenerateAddress()
				data, _ := erc20.Pack("", account)
				return data, &contract
			},
			false,
>>>>>>> main
		},
		{
			"pass",
			types.ModuleAddress,
<<<<<<< HEAD
			func() []byte {
				account := utiltx.GenerateAddress()
				data, _ := erc20.Pack("balanceOf", account)
				return data
			},
			false,
=======
			func() ([]byte, *common.Address) {
				contract, err := suite.DeployContract("coin", "token", erc20Decimals)
				suite.Require().NoError(err)
				account := utiltx.GenerateAddress()
				data, _ := erc20.Pack("balanceOf", account)
				return data, &contract
			},
>>>>>>> main
			true,
		},
		{
			"fail empty data",
			types.ModuleAddress,
<<<<<<< HEAD
			func() []byte {
				return []byte{}
			},
			false,
			false,
=======
			func() ([]byte, *common.Address) {
				contract, err := suite.DeployContract("coin", "token", erc20Decimals)
				suite.Require().NoError(err)
				return []byte{}, &contract
			},
			false,
>>>>>>> main
		},

		{
			"fail empty sender",
			common.Address{},
<<<<<<< HEAD
			func() []byte {
				return []byte{}
			},
			false,
			false,
=======
			func() ([]byte, *common.Address) {
				contract, err := suite.DeployContract("coin", "token", erc20Decimals)
				suite.Require().NoError(err)
				return []byte{}, &contract
			},
			false,
>>>>>>> main
		},
		{
			"deploy",
			types.ModuleAddress,
<<<<<<< HEAD
			func() []byte {
				ctorArgs, _ := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("", "test", "test", uint8(18))
				data := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...) //nolint:gocritic
				return data
			},
			true,
			true,
=======
			func() ([]byte, *common.Address) {
				ctorArgs, _ := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("", "test", "test", uint8(18))
				data := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...) //nolint:gocritic
				return data, nil
			},
			true,
>>>>>>> main
		},
		{
			"fail deploy",
			types.ModuleAddress,
<<<<<<< HEAD
			func() []byte {
				params := suite.network.App.EvmKeeper.GetParams(suite.network.GetContext())
				params.AccessControl.Create = evmtypes.AccessControlType{
					AccessType: evmtypes.AccessTypeRestricted,
				}
				_ = suite.network.App.EvmKeeper.SetParams(suite.network.GetContext(), params)
				ctorArgs, _ := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("", "test", "test", uint8(18))
				data := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...) //nolint:gocritic
				return data
			},
			true,
=======
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
>>>>>>> main
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

<<<<<<< HEAD
			data := tc.malleate()
			var res *evmtypes.MsgEthereumTxResponse
			var err error

			if tc.deploy {
				res, err = suite.network.App.EvmKeeper.CallEVMWithData(suite.network.GetContext(), tc.from, nil, data, true)
			} else {
				res, err = suite.network.App.EvmKeeper.CallEVMWithData(suite.network.GetContext(), tc.from, &wevmosContract, data, false)
			}

=======
			data, contract := tc.malleate()

			res, err := suite.app.EvmKeeper.CallEVMWithData(suite.ctx, tc.from, contract, data, true)
>>>>>>> main
			if tc.expPass {
				suite.Require().IsTypef(&evmtypes.MsgEthereumTxResponse{}, res, tc.name)
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
