package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/ethereum/go-ethereum/common"
	testfactory "github.com/evmos/evmos/v19/testutil/integration/evmos/factory"
	utiltx "github.com/evmos/evmos/v19/testutil/tx"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
	"github.com/stretchr/testify/mock"

	"github.com/evmos/evmos/v19/contracts"
	"github.com/evmos/evmos/v19/x/erc20/keeper"
	"github.com/evmos/evmos/v19/x/erc20/types"
	erc20mocks "github.com/evmos/evmos/v19/x/erc20/types/mocks"
)

func (suite *KeeperTestSuite) TestQueryERC20() {
	var (
		contract common.Address
		ctx      sdk.Context
	)
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
			func() {
				var err error
				contract, err = suite.factory.DeployContract(
					suite.keyring.GetPrivKey(0),
					evmtypes.EvmTxArgs{},
					testfactory.ContractDeploymentData{
						Contract:        contracts.ERC20MinterBurnerDecimalsContract,
						ConstructorArgs: []interface{}{"coin", "token", erc20Decimals},
					},
				)
				suite.Require().NoError(err)
				suite.Require().NoError(suite.network.NextBlock())
				ctx = suite.network.GetContext()
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset
		ctx = suite.network.GetContext()

		tc.malleate()

		res, err := suite.network.App.Erc20Keeper.QueryERC20(ctx, contract)
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
	var mockEVMKeeper *erc20mocks.EVMKeeper
	contract := utiltx.GenerateAddress()
	testCases := []struct {
		name       string
		malleate   func()
		expBalance int64
		res        bool
	}{
		{
			"Failed to call Evm",
			func() {
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
			},
			int64(0),
			false,
		},
		{
			"Incorrect res",
			func() {
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{0, 0}}, nil).Once()
			},
			int64(0),
			false,
		},
		{
			"Correct Execution",
			func() {
				balance := make([]uint8, 32)
				balance[31] = uint8(10)
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
			},
			int64(10),
			true,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset
		mockEVMKeeper = &erc20mocks.EVMKeeper{}
		suite.network.App.Erc20Keeper = keeper.NewKeeper(
			suite.network.App.GetKey("erc20"), suite.network.App.AppCodec(),
			authtypes.NewModuleAddress(govtypes.ModuleName),
			suite.network.App.AccountKeeper, suite.network.App.BankKeeper,
			mockEVMKeeper, suite.network.App.StakingKeeper,
			suite.network.App.AuthzKeeper, &suite.network.App.TransferKeeper,
		)

		tc.malleate()

		abi := contracts.ERC20MinterBurnerDecimalsContract.ABI
		balance := suite.network.App.Erc20Keeper.BalanceOf(suite.network.GetContext(), abi, contract, utiltx.GenerateAddress())
		if tc.res {
			suite.Require().Equal(balance.Int64(), tc.expBalance)
		} else {
			suite.Require().Nil(balance)
		}
	}
}

func (suite *KeeperTestSuite) TestQueryERC20ForceFail() {
	var mockEVMKeeper *erc20mocks.EVMKeeper
	contract := utiltx.GenerateAddress()
	testCases := []struct {
		name     string
		malleate func()
		res      bool
	}{
		{
			"Failed to call Evm",
			func() {
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
			},
			false,
		},
		{
			"Incorrect res",
			func() {
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{0, 0}}, nil).Once()
			},
			false,
		},
		{
			"Correct res for name - incorrect for symbol",
			func() {
				ret := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 67, 111, 105, 110, 32, 84, 111, 107, 101, 110, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: ret}, nil).Once()
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{VmError: "Error"}, nil).Once()
			},
			false,
		},
		{
			"incorrect symbol res",
			func() {
				ret := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 67, 111, 105, 110, 32, 84, 111, 107, 101, 110, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: ret}, nil).Once()
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{0, 0}}, nil).Once()
			},
			false,
		},
		{
			"Correct res for name - incorrect for symbol",
			func() {
				ret := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 67, 111, 105, 110, 32, 84, 111, 107, 101, 110, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
				retSymbol := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 67, 84, 75, 78, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: ret}, nil).Once()
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: retSymbol}, nil).Once()
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{VmError: "Error"}, nil).Once()
			},
			false,
		},
		{
			"incorrect symbol res",
			func() {
				ret := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 67, 111, 105, 110, 32, 84, 111, 107, 101, 110, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
				retSymbol := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 67, 84, 75, 78, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: ret}, nil).Once()
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: retSymbol}, nil).Once()
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{0, 0}}, nil).Once()
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset

		// TODO: what's the reason we are using mockEVMKeeper here? Instead of just passing the suite.app.EvmKeeper?
		mockEVMKeeper = &erc20mocks.EVMKeeper{}
		suite.network.App.Erc20Keeper = keeper.NewKeeper(
			suite.network.App.GetKey("erc20"), suite.network.App.AppCodec(),
			authtypes.NewModuleAddress(govtypes.ModuleName), suite.network.App.AccountKeeper,
			suite.network.App.BankKeeper, mockEVMKeeper, suite.network.App.StakingKeeper,
			suite.network.App.AuthzKeeper, &suite.network.App.TransferKeeper,
		)

		tc.malleate()

		res, err := suite.network.App.Erc20Keeper.QueryERC20(suite.network.GetContext(), contract)
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
