package keeper_test

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"
	testutils "github.com/evmos/evmos/v20/testutil/integration/evmos/utils"
	utiltx "github.com/evmos/evmos/v20/testutil/tx"
	"github.com/evmos/evmos/v20/x/erc20/keeper"
	"github.com/evmos/evmos/v20/x/erc20/types"
	erc20mocks "github.com/evmos/evmos/v20/x/erc20/types/mocks"
	"github.com/evmos/evmos/v20/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
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
				mockBankKeeper.EXPECT().GetBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(sdk.Coin{Denom: cosmosTokenDisplay, Amount: math.OneInt()}).AnyTimes()
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
				mockBankKeeper.EXPECT().GetBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(sdk.Coin{Denom: cosmosTokenDisplay, Amount: math.OneInt()})
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

func (suite *KeeperTestSuite) TestMint() {
	var ctx sdk.Context
	contractAddr := utiltx.GenerateAddress()
	denom := cosmosTokenDisplay
	sender := sdk.AccAddress(utiltx.GenerateAddress().Bytes())

	testcases := []struct {
		name     string
		msgMint  *types.MsgMint
		malleate func()
		expPass  bool
	}{
		{
			"fail - invalid sender address",
			&types.MsgMint{
				ContractAddress: utiltx.GenerateAddress().String(),
				Amount:          math.NewInt(100),
				Sender:          "invalid",
			},
			func() {},
			false,
		},
		{
			"fail - invalid receiver address",
			&types.MsgMint{
				ContractAddress: contractAddr.String(),
				Amount:          math.NewInt(100),
				Sender:          sender.String(),
				To:              "invalid",
			},
			func() {},
			false,
		},
		{
			"pass - valid msg",
			&types.MsgMint{
				ContractAddress: contractAddr.String(),
				Amount:          math.NewInt(100),
				Sender:          sender.String(),
				To:              sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
			},
			func() {
				expPair := types.NewTokenPair(
					contractAddr,
					denom,
					types.OWNER_MODULE,
				)
				expPair.SetOwnerAddress(sender.String())
				suite.network.App.Erc20Keeper.SetTokenPair(ctx, expPair)
				suite.network.App.Erc20Keeper.SetDenomMap(ctx, expPair.Denom, expPair.GetID())
				suite.network.App.Erc20Keeper.SetERC20Map(ctx, expPair.GetERC20Contract(), expPair.GetID())
			},
			true,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			ctx = suite.network.GetContext()
			tc.malleate()
			res, err := suite.network.App.Erc20Keeper.Mint(ctx, tc.msgMint)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestBurn() {
	var ctx sdk.Context
	contractAddr := utiltx.GenerateAddress()
	sender := sdk.AccAddress(utiltx.GenerateAddress().Bytes())

	testcases := []struct {
		name     string
		msgBurn  *types.MsgBurn
		malleate func()
		expPass  bool
	}{
		{
			"fail - invalid sender address",
			&types.MsgBurn{
				ContractAddress: contractAddr.String(),
				Amount:          math.NewInt(100),
				Sender:          "invalid",
			},
			func() {},
			false,
		},
		{
			"pass - valid msg",
			&types.MsgBurn{
				ContractAddress: contractAddr.String(),
				Amount:          math.NewInt(100),
				Sender:          sender.String(),
			},
			func() {
				expPair := types.NewTokenPair(
					contractAddr,
					cosmosTokenDisplay,
					types.OWNER_MODULE,
				)
				suite.network.App.Erc20Keeper.SetTokenPair(ctx, expPair)
				suite.network.App.Erc20Keeper.SetDenomMap(ctx, expPair.Denom, expPair.GetID())
				suite.network.App.Erc20Keeper.SetERC20Map(ctx, expPair.GetERC20Contract(), expPair.GetID())
			},
			true,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			ctx = suite.network.GetContext()

			err := suite.network.App.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.Coins{sdk.NewCoin(cosmosTokenDisplay, math.NewInt(100))})
			suite.Require().NoError(err)

			err = suite.network.App.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, sdk.Coins{sdk.NewCoin(cosmosTokenDisplay, math.NewInt(100))})
			suite.Require().NoError(err)

			tc.malleate()
			res, err := suite.network.App.Erc20Keeper.Burn(ctx, tc.msgBurn)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestTransferContractOwnership() {
	var ctx sdk.Context

	tokenAddr := utiltx.GenerateAddress()

	testcases := []struct {
		name     string
		msg      *types.MsgTransferOwnership
		malleate func()
		expPass  bool
	}{
		{
			"fail - invalid authority address",
			&types.MsgTransferOwnership{
				Authority: "invalid",
			},
			func() {},
			false,
		},
		{
			"fail - invalid new owner address",
			&types.MsgTransferOwnership{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				NewOwner:  "invalid",
			},
			func() {},
			false,
		},
		{
			"pass - valid msg",
			&types.MsgTransferOwnership{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				NewOwner:  sdk.AccAddress(utiltx.GenerateAddress().Bytes()).String(),
				Token:     tokenAddr.String(),
			},
			func() {
				expPair := types.NewTokenPair(
					tokenAddr,
					cosmosTokenDisplay,
					types.OWNER_MODULE,
				)
				suite.network.App.Erc20Keeper.SetTokenPair(ctx, expPair)
				suite.network.App.Erc20Keeper.SetDenomMap(ctx, expPair.Denom, expPair.GetID())
				suite.network.App.Erc20Keeper.SetERC20Map(ctx, expPair.GetERC20Contract(), expPair.GetID())
			},
			true,
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			ctx = sdk.UnwrapSDKContext(suite.network.GetContext())
			tc.malleate()
			res, err := suite.network.App.Erc20Keeper.TransferContractOwnership(ctx, tc.msg)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
