package keeper_test

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"
	testutils "github.com/evmos/evmos/v19/testutil/integration/evmos/utils"
	"github.com/evmos/evmos/v19/x/erc20/keeper"
	"github.com/evmos/evmos/v19/x/erc20/types"
	erc20mocks "github.com/evmos/evmos/v19/x/erc20/types/mocks"
	"github.com/evmos/evmos/v19/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
)

func (suite *KeeperTestSuite) TestConvertERC20NativeERC20() {
	var (
		contractAddr common.Address
		coinName     string
	)
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
				err := testutils.UpdateERC20Params(
					testutils.UpdateParamsInput{
						Tf:      suite.factory,
						Network: suite.network,
						Pk:      suite.keyring.GetPrivKey(0),
						Params:  params,
					},
				)
				suite.Require().NoError(err)
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
			"fail - force evm fail",
			100,
			10,
			func(common.Address) {},
			func() {
				mockEVMKeeper := &erc20mocks.EVMKeeper{}
				suite.network.App.Erc20Keeper = keeper.NewKeeper(
					suite.network.App.GetKey("erc20"), suite.network.App.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.network.App.AccountKeeper,
					suite.network.App.BankKeeper, mockEVMKeeper, suite.network.App.StakingKeeper,
					suite.network.App.AuthzKeeper, &suite.network.App.TransferKeeper,
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
				suite.network.App.Erc20Keeper = keeper.NewKeeper(
					suite.network.App.GetKey("erc20"), suite.network.App.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.network.App.AccountKeeper,
					suite.network.App.BankKeeper, mockEVMKeeper, suite.network.App.StakingKeeper,
					suite.network.App.AuthzKeeper, &suite.network.App.TransferKeeper,
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
				suite.network.App.Erc20Keeper = keeper.NewKeeper(
					suite.network.App.GetKey("erc20"), suite.network.App.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.network.App.AccountKeeper,
					suite.network.App.BankKeeper, mockEVMKeeper, suite.network.App.StakingKeeper,
					suite.network.App.AuthzKeeper, &suite.network.App.TransferKeeper,
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
				suite.network.App.Erc20Keeper = keeper.NewKeeper(
					suite.network.App.GetKey("erc20"), suite.network.App.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.network.App.AccountKeeper,
					suite.network.App.BankKeeper, mockEVMKeeper, suite.network.App.StakingKeeper,
					suite.network.App.AuthzKeeper, &suite.network.App.TransferKeeper,
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
				ctrl := gomock.NewController(suite.T())
				mockBankKeeper := erc20mocks.NewMockBankKeeper(ctrl)

				suite.network.App.Erc20Keeper = keeper.NewKeeper(
					suite.network.App.GetKey("erc20"), suite.network.App.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.network.App.AccountKeeper,
					mockBankKeeper, suite.network.App.EvmKeeper, suite.network.App.StakingKeeper,
					suite.network.App.AuthzKeeper, &suite.network.App.TransferKeeper,
				)

				mockBankKeeper.EXPECT().MintCoins(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("failed to mint")).AnyTimes()
				mockBankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("failed to unescrow")).AnyTimes()
				mockBankKeeper.EXPECT().BlockedAddr(gomock.Any()).Return(false).AnyTimes()
				mockBankKeeper.EXPECT().GetBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(sdk.Coin{Denom: "coin", Amount: math.OneInt()}).AnyTimes()
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
				ctrl := gomock.NewController(suite.T())
				mockBankKeeper := erc20mocks.NewMockBankKeeper(ctrl)

				suite.network.App.Erc20Keeper = keeper.NewKeeper(
					suite.network.App.GetKey("erc20"), suite.network.App.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.network.App.AccountKeeper,
					mockBankKeeper, suite.network.App.EvmKeeper, suite.network.App.StakingKeeper,
					suite.network.App.AuthzKeeper, &suite.network.App.TransferKeeper,
				)

				mockBankKeeper.EXPECT().MintCoins(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				mockBankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("failed to unescrow"))
				mockBankKeeper.EXPECT().BlockedAddr(gomock.Any()).Return(false)
				mockBankKeeper.EXPECT().GetBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(sdk.Coin{Denom: "coin", Amount: math.OneInt()})
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
				ctrl := gomock.NewController(suite.T())
				mockBankKeeper := erc20mocks.NewMockBankKeeper(ctrl)

				suite.network.App.Erc20Keeper = keeper.NewKeeper(
					suite.network.App.GetKey("erc20"), suite.network.App.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.network.App.AccountKeeper,
					mockBankKeeper, suite.network.App.EvmKeeper, suite.network.App.StakingKeeper,
					suite.network.App.AuthzKeeper, &suite.network.App.TransferKeeper,
				)

				mockBankKeeper.EXPECT().MintCoins(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				mockBankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				mockBankKeeper.EXPECT().BlockedAddr(gomock.Any()).Return(false)
				mockBankKeeper.EXPECT().GetBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(sdk.Coin{Denom: coinName, Amount: math.OneInt()}).AnyTimes()
			},
			contractMinterBurner,
			false,
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			var err error
			suite.mintFeeCollector = true
			defer func() {
				suite.mintFeeCollector = false
			}()

			suite.SetupTest()

			contractAddr, err = suite.setupRegisterERC20Pair(tc.contractType)
			suite.Require().NoError(err)

			tc.malleate(contractAddr)
			suite.Require().NotNil(contractAddr)

			coinName = types.CreateDenom(contractAddr.String())
			sender := suite.keyring.GetAccAddr(0)

			_, err = suite.MintERC20Token(contractAddr, suite.keyring.GetAddr(0), big.NewInt(tc.mint))
			suite.Require().NoError(err)
			// update context with latest committed changes

			tc.extra()

			msg := types.NewMsgConvertERC20(
				math.NewInt(tc.transfer),
				sender,
				contractAddr,
				suite.keyring.GetAddr(0),
			)

			ctx := suite.network.GetContext()
			_, err = suite.network.App.Erc20Keeper.ConvertERC20(ctx, msg)

			cosmosBalance := suite.network.App.BankKeeper.GetBalance(ctx, sender, coinName)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)

				acc := suite.network.App.EvmKeeper.GetAccountWithoutBalance(ctx, contractAddr)
				if tc.selfdestructed {
					suite.Require().Nil(acc, "expected contract to be destroyed")
				} else {
					suite.Require().NotNil(acc)
				}

				if tc.selfdestructed || !acc.IsContract() {
					id := suite.network.App.Erc20Keeper.GetTokenPairID(ctx, contractAddr.String())
					_, found := suite.network.App.Erc20Keeper.GetTokenPair(ctx, id)
					suite.Require().False(found)
				} else {
					suite.Require().Equal(cosmosBalance.Amount, math.NewInt(tc.transfer))
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
		suite.Run("MsgUpdateParams", func() {
			suite.SetupTest()
			_, err := suite.network.App.Erc20Keeper.UpdateParams(suite.network.GetContext(), tc.request)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
