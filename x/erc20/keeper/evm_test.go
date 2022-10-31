package keeper_test

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/tests"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/mock"

	"github.com/evmos/evmos/v10/contracts"
	"github.com/evmos/evmos/v10/x/erc20/keeper"
	"github.com/evmos/evmos/v10/x/erc20/types"
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

func (suite *KeeperTestSuite) TestBalanceOf() {
	var mockEVMKeeper *MockEVMKeeper
	contract := tests.GenerateAddress()
	testCases := []struct {
		name       string
		malleate   func()
		expBalance int64
		res        bool
	}{
		{
			"Failed to call Evm",
			func() {
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
			},
			int64(0),
			false,
		},
		{
			"Incorrect res",
			func() {
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{0, 0}}, nil).Once()
			},
			int64(0),
			false,
		},
		{
			"Correct Execution",
			func() {
				balance := make([]uint8, 32)
				balance[31] = uint8(10)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
			},
			int64(10),
			true,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset
		mockEVMKeeper = &MockEVMKeeper{}
		sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
		suite.Require().True(found)
		suite.app.Erc20Keeper = keeper.NewKeeper(suite.app.GetKey("erc20"), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper)

		tc.malleate()

		abi := contracts.ERC20BurnableContract.ABI
		balance := suite.app.Erc20Keeper.BalanceOf(suite.ctx, abi, contract, tests.GenerateAddress())
		if tc.res {
			suite.Require().Equal(balance.Int64(), tc.expBalance)
		} else {
			suite.Require().Nil(balance)
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

func (suite *KeeperTestSuite) TestForceFail() {
	var mockEVMKeeper *MockEVMKeeper
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
	testCases := []struct {
		name     string
		malleate func()
		commit   bool
		expPass  bool
	}{
		{
			"Force estimate gas error",
			func() {
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced EstimateGas error"))
			},
			true,
			false,
		},
		{
			"Force ApplyMessage error",
			func() {
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
			},
			true,
			false,
		},
		{
			"Force ApplyMessage failed",
			func() {
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{VmError: "SomeError"}, nil)
			},
			true,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			mockEVMKeeper = &MockEVMKeeper{}
			sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
			suite.Require().True(found)
			suite.app.Erc20Keeper = keeper.NewKeeper(suite.app.GetKey("erc20"), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper)

			tc.malleate()

			contract, err := suite.DeployContract("coin", "token", erc20Decimals)
			suite.Require().NoError(err)
			account := tests.GenerateAddress()
			data, _ := erc20.Pack("balanceOf", account)

			res, err := suite.app.Erc20Keeper.CallEVMWithData(suite.ctx, types.ModuleAddress, &contract, data, tc.commit)
			if tc.expPass {
				suite.Require().IsTypef(&evmtypes.MsgEthereumTxResponse{}, res, tc.name)
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryERC20ForceFail() {
	var mockEVMKeeper *MockEVMKeeper
	contract := tests.GenerateAddress()
	testCases := []struct {
		name     string
		malleate func()
		res      bool
	}{
		{
			"Failed to call Evm",
			func() {
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
			},
			false,
		},
		{
			"Incorrect res",
			func() {
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{0, 0}}, nil).Once()
			},
			false,
		},
		{
			"Correct res for name - incorrect for symbol",
			func() {
				ret := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 67, 111, 105, 110, 32, 84, 111, 107, 101, 110, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: ret}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{VmError: "Error"}, nil).Once()
			},
			false,
		},
		{
			"incorrect symbol res",
			func() {
				ret := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 67, 111, 105, 110, 32, 84, 111, 107, 101, 110, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: ret}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{0, 0}}, nil).Once()
			},
			false,
		},
		{
			"Correct res for name - incorrect for symbol",
			func() {
				ret := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 67, 111, 105, 110, 32, 84, 111, 107, 101, 110, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
				retSymbol := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 67, 84, 75, 78, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: ret}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: retSymbol}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{VmError: "Error"}, nil).Once()
			},
			false,
		},
		{
			"incorrect symbol res",
			func() {
				ret := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 67, 111, 105, 110, 32, 84, 111, 107, 101, 110, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
				retSymbol := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 67, 84, 75, 78, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: ret}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: retSymbol}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{0, 0}}, nil).Once()
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset
		mockEVMKeeper = &MockEVMKeeper{}
		sp, found := suite.app.ParamsKeeper.GetSubspace(types.ModuleName)
		suite.Require().True(found)
		suite.app.Erc20Keeper = keeper.NewKeeper(suite.app.GetKey("erc20"), suite.app.AppCodec(), sp, suite.app.AccountKeeper, suite.app.BankKeeper, mockEVMKeeper)

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
