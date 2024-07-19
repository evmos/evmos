package keeper_test

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/x/erc20/keeper"
	"github.com/evmos/evmos/v19/x/erc20/types"
	erc20mocks "github.com/evmos/evmos/v19/x/erc20/types/mocks"
	"github.com/evmos/evmos/v19/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
	"github.com/stretchr/testify/mock"
)

func (suite *KeeperTestSuite) TestConvertERC20NativeERC20() {
	var contractAddr common.Address
	var coinName string

	testCases := []struct {
		name           string
		mint           int64
		transfer       int64
		malleate       func(common.Address)
		extra          func()
		contractType   int
		expPass        bool
		selfdestructed bool
	}{
		{
			"ok - sufficient funds",
			100,
			10,
			func(common.Address) {},
			func() {},
			contractMinterBurner,
			true,
			false,
		},
		{
			"ok - equal funds",
			10,
			10,
			func(common.Address) {},
			func() {},
			contractMinterBurner,
			true,
			false,
		},
		{
			"ok - equal funds",
			10,
			10,
			func(common.Address) {},
			func() {},
			contractMinterBurner,
			true,
			false,
		},
		{
			"ok - suicided contract",
			10,
			10,
			func(erc20 common.Address) {
				stateDB := suite.StateDB()
				ok := stateDB.Suicide(erc20)
				suite.Require().True(ok)
				suite.Require().NoError(stateDB.Commit())
			},
			func() {},
			contractMinterBurner,
			true,
			true,
		},
		{
			"fail - insufficient funds - callEVM",
			0,
			10,
			func(common.Address) {},
			func() {},
			contractMinterBurner,
			false,
			false,
		},
		{
			"fail - minting disabled",
			100,
			10,
			func(common.Address) {
				params := types.DefaultParams()
				params.EnableErc20 = false
				suite.app.Erc20Keeper.SetParams(suite.ctx, params) //nolint:errcheck
			},
			func() {},
			contractMinterBurner,
			false,
			false,
		},
		{
			"fail - direct balance manipulation contract",
			100,
			10,
			func(common.Address) {},
			func() {},
			contractDirectBalanceManipulation,
			false,
			false,
		},
		{
			"fail - delayed malicious contract",
			10,
			10,
			func(common.Address) {},
			func() {},
			contractMaliciousDelayed,
			false,
			false,
		},
		{
			"fail - negative transfer contract",
			10,
			-10,
			func(common.Address) {},
			func() {},
			contractMinterBurner,
			false,
			false,
		},
		{
			"fail - no module address",
			100,
			10,
			func(common.Address) {
			},
			func() {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, types.ModuleAddress.Bytes())
				suite.app.AccountKeeper.RemoveAccount(suite.ctx, acc)
			},
			contractMinterBurner,
			false,
			false,
		},
		{
			"fail - force evm fail",
			100,
			10,
			func(common.Address) {},
			func() {
				mockEVMKeeper := &erc20mocks.EVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper,
					suite.app.AuthzKeeper, &suite.app.TransferKeeper,
				)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGasInternal", mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("CallEVMWithData", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			contractMinterBurner,
			false,
			false,
		},
		{
			"fail - force get balance fail",
			100,
			10,
			func(common.Address) {},
			func() {
				mockEVMKeeper := &erc20mocks.EVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper,
					suite.app.AuthzKeeper, &suite.app.TransferKeeper,
				)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				balance[31] = uint8(1)
				mockEVMKeeper.On("EstimateGasInternal", mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Twice()
				mockEVMKeeper.On("CallEVMWithData", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced balance error"))
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			contractMinterBurner,
			false,
			false,
		},
		{
			"fail - force transfer unpack fail",
			100,
			10,
			func(common.Address) {},
			func() {
				mockEVMKeeper := &erc20mocks.EVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper,
					suite.app.AuthzKeeper, &suite.app.TransferKeeper,
				)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGasInternal", mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("CallEVMWithData", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			contractMinterBurner,
			false,
			false,
		},

		{
			"fail - force invalid transfer fail",
			100,
			10,
			func(common.Address) {},
			func() {
				mockEVMKeeper := &erc20mocks.EVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper,
					suite.app.AuthzKeeper, &suite.app.TransferKeeper,
				)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGasInternal", mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("CallEVMWithData", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			contractMinterBurner,
			false,
			false,
		},
		{
			"fail - force mint fail",
			100,
			10,
			func(common.Address) {},
			func() {
				mockBankKeeper := &erc20mocks.BankKeeper{}

				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					mockBankKeeper, suite.app.EvmKeeper, suite.app.StakingKeeper,
					suite.app.AuthzKeeper, &suite.app.TransferKeeper,
				)

				mockBankKeeper.On("MintCoins", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to mint"))
				mockBankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to unescrow"))
				mockBankKeeper.On("BlockedAddr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)
				mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sdk.Coin{Denom: "coin", Amount: math.OneInt()})
			},
			contractMinterBurner,
			false,
			false,
		},
		{
			"fail - force send minted fail",
			100,
			10,
			func(common.Address) {},
			func() {
				mockBankKeeper := &erc20mocks.BankKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					mockBankKeeper, suite.app.EvmKeeper, suite.app.StakingKeeper,
					suite.app.AuthzKeeper, &suite.app.TransferKeeper,
				)

				mockBankKeeper.On("MintCoins", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockBankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to unescrow"))
				mockBankKeeper.On("BlockedAddr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)
				mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sdk.Coin{Denom: "coin", Amount: math.OneInt()})
			},
			contractMinterBurner,
			false,
			false,
		},
		{
			"fail - force bank balance fail",
			100,
			10,
			func(common.Address) {},
			func() {
				mockBankKeeper := &erc20mocks.BankKeeper{}

				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					mockBankKeeper, suite.app.EvmKeeper, suite.app.StakingKeeper,
					suite.app.AuthzKeeper, &suite.app.TransferKeeper,
				)

				mockBankKeeper.On("MintCoins", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockBankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockBankKeeper.On("BlockedAddr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)
				mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sdk.Coin{Denom: coinName, Amount: math.NewInt(int64(10))})
			},
			contractMinterBurner,
			false,
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()

			contractAddr = suite.setupRegisterERC20Pair(tc.contractType)

			tc.malleate(contractAddr)
			suite.Require().NotNil(contractAddr)
			suite.Commit()

			coinName = types.CreateDenom(contractAddr.String())
			sender := sdk.AccAddress(suite.address.Bytes())
			msg := types.NewMsgConvertERC20(
				math.NewInt(tc.transfer),
				sender,
				contractAddr,
				suite.address,
			)

			suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(tc.mint))
			suite.Commit()
			ctx := sdk.WrapSDKContext(suite.ctx)

			tc.extra()
			res, err := suite.app.Erc20Keeper.ConvertERC20(ctx, msg)

			expRes := &types.MsgConvertERC20Response{}
			suite.Commit()
			balance := suite.BalanceOf(contractAddr, suite.address)
			cosmosBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sender, coinName)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)

				acc := suite.app.EvmKeeper.GetAccountWithoutBalance(suite.ctx, contractAddr)
				if tc.selfdestructed {
					suite.Require().Nil(acc, "expected contract to be destroyed")
				} else {
					suite.Require().NotNil(acc)
				}

				if tc.selfdestructed || !acc.IsContract() {
					id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
					_, found := suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
					suite.Require().False(found)
				} else {
					suite.Require().Equal(expRes, res)
					suite.Require().Equal(cosmosBalance.Amount, math.NewInt(tc.transfer))
					suite.Require().Equal(balance.(*big.Int).Int64(), big.NewInt(tc.mint-tc.transfer).Int64())
				}
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
	suite.mintFeeCollector = false
}

func (suite *KeeperTestSuite) TestUpdateParams() {
	testCases := []struct {
		name      string
		request   *types.MsgUpdateParams
		expectErr bool
	}{
		{
			name:      "fail - invalid authority",
			request:   &types.MsgUpdateParams{Authority: "foobar"},
			expectErr: true,
		},
		{
			name: "pass - valid Update msg",
			request: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params:    types.DefaultParams(),
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run("MsgUpdateParams", func() {
			_, err := suite.app.Erc20Keeper.UpdateParams(suite.ctx, tc.request)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
