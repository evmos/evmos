package keeper_test

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/contracts"
	utiltx "github.com/evmos/evmos/v20/testutil/tx"
	"github.com/evmos/evmos/v20/x/erc20/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

func (suite *KeeperTestSuite) TestCallEVM() {
	wevmosContract := common.HexToAddress(types.WEVMOSContractMainnet)
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
		account := utiltx.GenerateAddress()
		res, err := suite.network.App.EvmKeeper.CallEVM(suite.network.GetContext(), erc20, types.ModuleAddress, wevmosContract, false, tc.method, account)
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
	wevmosContract := common.HexToAddress(types.WEVMOSContractMainnet)
	testCases := []struct {
		name     string
		from     common.Address
		malleate func() []byte
		deploy   bool
		expPass  bool
	}{
		{
			"pass with unknown method",
			types.ModuleAddress,
			func() []byte {
				account := utiltx.GenerateAddress()
				data, _ := erc20.Pack("", account)
				return data
			},
			false,
			true,
		},
		{
			"pass",
			types.ModuleAddress,
			func() []byte {
				account := utiltx.GenerateAddress()
				data, _ := erc20.Pack("balanceOf", account)
				return data
			},
			false,
			true,
		},
		{
			"pass with empty data",
			types.ModuleAddress,
			func() []byte {
				return []byte{}
			},
			false,
			true,
		},

		{
			"fail empty sender",
			common.Address{},
			func() []byte {
				return []byte{}
			},
			false,
			false,
		},
		{
			"deploy",
			types.ModuleAddress,
			func() []byte {
				ctorArgs, _ := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("", "test", "test", uint8(18))
				data := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...) //nolint:gocritic
				return data
			},
			true,
			true,
		},
		{
			"fail deploy",
			types.ModuleAddress,
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
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			data := tc.malleate()
			var res *evmtypes.MsgEthereumTxResponse
			var err error

			if tc.deploy {
				res, err = suite.network.App.EvmKeeper.CallEVMWithData(suite.network.GetContext(), tc.from, nil, data, true)
			} else {
				res, err = suite.network.App.EvmKeeper.CallEVMWithData(suite.network.GetContext(), tc.from, &wevmosContract, data, false)
			}

			if tc.expPass {
				suite.Require().IsTypef(&evmtypes.MsgEthereumTxResponse{}, res, tc.name)
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
