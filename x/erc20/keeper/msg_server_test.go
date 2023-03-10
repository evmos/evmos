package keeper_test

import (
	"fmt"
	"math/big"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/mock"

	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v12/x/erc20/keeper"
	"github.com/evmos/evmos/v12/x/erc20/types"
	"github.com/evmos/evmos/v12/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v12/x/evm/types"
)

func (suite *KeeperTestSuite) TestConvertCoinNativeCoin() {
	testCases := []struct { //nolint:dupl
		name           string
		mint           int64
		burn           int64
		malleate       func(common.Address)
		extra          func()
		expPass        bool
		selfdestructed bool
	}{
		{
			"ok - sufficient funds",
			100,
			10,
			func(common.Address) {},
			func() {},
			true,
			false,
		},
		{
			"ok - equal funds",
			10,
			10,
			func(common.Address) {},
			func() {},
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
			true,
			true,
		},
		{
			"fail - insufficient funds",
			0,
			10,
			func(common.Address) {},
			func() {},
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
			false,
			false,
		},
		{
			"fail - deleted module account - force fail", 100, 10, func(common.Address) {},
			func() {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, types.ModuleAddress.Bytes())
				suite.app.AccountKeeper.RemoveAccount(suite.ctx, acc)
			}, false, false,
		},
		{ //nolint:dupl
			"fail - force evm fail", 100, 10, func(common.Address) {},
			func() { //nolint:dupl
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			}, false, false,
		},
		{
			"fail - force evm balance error", 100, 10, func(common.Address) {},
			func() { //nolint:dupl
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				// first balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				// convert coin
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil).Once()
				// second balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, fmt.Errorf("third")).Once()
				// Extra call on test
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			}, false, false,
		},
		{
			"fail - force balance error", 100, 10, func(common.Address) {},
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Times(4)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			}, false, false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			pair := suite.setupRegisterCoin(metadataCoin)
			suite.Require().NotNil(metadataCoin)
			erc20 := pair.GetERC20Contract()
			tc.malleate(erc20)
			suite.Commit()

			ctx := sdk.WrapSDKContext(suite.ctx)
			coins := sdk.NewCoins(sdk.NewCoin(cosmosTokenBase, sdk.NewInt(tc.mint)))
			sender := sdk.AccAddress(suite.address.Bytes())
			msg := types.NewMsgConvertCoin(
				sdk.NewCoin(cosmosTokenBase, sdk.NewInt(tc.burn)),
				suite.address,
				sender,
			)

			err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
			suite.Require().NoError(err, tc.name)
			err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)
			suite.Require().NoError(err, tc.name)

			tc.extra()
			res, err := suite.app.Erc20Keeper.ConvertCoin(ctx, msg)
			expRes := &types.MsgConvertCoinResponse{}
			suite.Commit()
			balance := suite.BalanceOf(common.HexToAddress(pair.Erc20Address), suite.address)
			cosmosBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sender, metadataCoin.Base)

			if tc.expPass {
				suite.Require().NoError(err, tc.name)

				acc := suite.app.EvmKeeper.GetAccountWithoutBalance(suite.ctx, erc20)
				if tc.selfdestructed {
					suite.Require().Nil(acc, "expected contract to be destroyed")
				} else {
					suite.Require().NotNil(acc)
				}

				if tc.selfdestructed || !acc.IsContract() {
					id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, erc20.String())
					_, found := suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
					suite.Require().False(found)
				} else {
					suite.Require().Equal(expRes, res)
					suite.Require().Equal(cosmosBalance.Amount.Int64(), sdk.NewInt(tc.mint-tc.burn).Int64())
					suite.Require().Equal(balance.(*big.Int).Int64(), big.NewInt(tc.burn).Int64())
				}
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
	suite.mintFeeCollector = false
}

func (suite *KeeperTestSuite) TestConvertERC20NativeCoin() {
	testCases := []struct {
		name      string
		mint      int64
		burn      int64
		reconvert int64
		malleate  func()
		expPass   bool
	}{
		{"ok - sufficient funds", 100, 10, 5, func() {}, true},
		{"ok - equal funds", 10, 10, 10, func() {}, true},
		{"fail - insufficient funds", 10, 1, 5, func() {}, false},
		{"fail ", 10, 1, -5, func() {}, false},
		{
			"fail - deleted module account - force fail", 100, 10, 5,
			func() {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, types.ModuleAddress.Bytes())
				suite.app.AccountKeeper.RemoveAccount(suite.ctx, acc)
			},
			false,
		},
		{ //nolint:dupl
			"fail - force evm fail", 100, 10, 5,
			func() { //nolint:dupl
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			false,
		},
		{
			"fail - force fail second balance", 100, 10, 5,
			func() { //nolint:dupl
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				// first balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				// convert coin
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil).Once()
				// second balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, fmt.Errorf("third")).Once()
				// Extra call on test
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			false,
		},
		{
			"fail - force fail second balance", 100, 10, 5,
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				// first balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				// convert coin
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil).Once()
				// second balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				// Extra call on test
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			false,
		},
		{
			"fail - force fail unescrow", 100, 10, 5,
			func() {
				mockBankKeeper := &MockBankKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					mockBankKeeper, suite.app.EvmKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				mockBankKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(nil)
				mockBankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to unescrow"))
				mockBankKeeper.On("BlockedAddr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)
				mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sdk.Coin{Denom: "coin", Amount: sdk.OneInt()})
			},
			false,
		},
		{
			"fail - force fail balance after transfer", 100, 10, 5,
			func() {
				mockBankKeeper := &MockBankKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					mockBankKeeper, suite.app.EvmKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				mockBankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockBankKeeper.On("BlockedAddr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)
				mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sdk.Coin{Denom: "acoin", Amount: sdk.OneInt()})
			},
			false,
		},
	}
	for _, tc := range testCases { //nolint:dupl
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			pair := suite.setupRegisterCoin(metadataCoin)
			suite.Require().NotNil(metadataCoin)
			suite.Require().NotNil(pair)

			// Precondition: Convert Coin to ERC20
			coins := sdk.NewCoins(sdk.NewCoin(cosmosTokenBase, sdk.NewInt(tc.mint)))
			sender := sdk.AccAddress(suite.address.Bytes())
			err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
			suite.Require().NoError(err, tc.name)
			err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)
			suite.Require().NoError(err, tc.name)
			msg := types.NewMsgConvertCoin(
				sdk.NewCoin(cosmosTokenBase, sdk.NewInt(tc.burn)),
				suite.address,
				sender,
			)

			ctx := sdk.WrapSDKContext(suite.ctx)
			_, err = suite.app.Erc20Keeper.ConvertCoin(ctx, msg)
			suite.Require().NoError(err, tc.name)
			suite.Commit()
			balance := suite.BalanceOf(common.HexToAddress(pair.Erc20Address), suite.address)
			cosmosBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sender, metadataCoin.Base)
			suite.Require().Equal(cosmosBalance.Amount.Int64(), sdk.NewInt(tc.mint-tc.burn).Int64())
			suite.Require().Equal(balance, big.NewInt(tc.burn))

			// Convert ERC20s back to Coins
			ctx = sdk.WrapSDKContext(suite.ctx)
			contractAddr := common.HexToAddress(pair.Erc20Address)
			msgConvertERC20 := types.NewMsgConvertERC20(
				sdk.NewInt(tc.reconvert),
				sender,
				contractAddr,
				suite.address,
			)

			tc.malleate()
			res, err := suite.app.Erc20Keeper.ConvertERC20(ctx, msgConvertERC20)
			expRes := &types.MsgConvertERC20Response{}
			suite.Commit()
			balance = suite.BalanceOf(contractAddr, suite.address)
			cosmosBalance = suite.app.BankKeeper.GetBalance(suite.ctx, sender, pair.Denom)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expRes, res)
				suite.Require().Equal(cosmosBalance.Amount.Int64(), sdk.NewInt(tc.mint-tc.burn+tc.reconvert).Int64())
				suite.Require().Equal(balance.(*big.Int).Int64(), big.NewInt(tc.burn-tc.reconvert).Int64())
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
	suite.mintFeeCollector = false
}

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
			func() { //nolint:dupl
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
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
			func() { //nolint:dupl
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				balance[31] = uint8(1)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Twice()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced balance error"))
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
			func() { //nolint:dupl
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil)
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
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil)
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
				mockBankKeeper := &MockBankKeeper{}

				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					mockBankKeeper, suite.app.EvmKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				mockBankKeeper.On("MintCoins", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to mint"))
				mockBankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to unescrow"))
				mockBankKeeper.On("BlockedAddr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)
				mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sdk.Coin{Denom: "coin", Amount: sdk.OneInt()})
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
				mockBankKeeper := &MockBankKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					mockBankKeeper, suite.app.EvmKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				mockBankKeeper.On("MintCoins", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockBankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to unescrow"))
				mockBankKeeper.On("BlockedAddr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)
				mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sdk.Coin{Denom: "coin", Amount: sdk.OneInt()})
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
				mockBankKeeper := &MockBankKeeper{}

				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					mockBankKeeper, suite.app.EvmKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				mockBankKeeper.On("MintCoins", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockBankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockBankKeeper.On("BlockedAddr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)
				mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sdk.Coin{Denom: coinName, Amount: sdk.NewInt(int64(10))})
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
				sdk.NewInt(tc.transfer),
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
					suite.Require().Equal(cosmosBalance.Amount, sdk.NewInt(tc.transfer))
					suite.Require().Equal(balance.(*big.Int).Int64(), big.NewInt(tc.mint-tc.transfer).Int64())
				}
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
	suite.mintFeeCollector = false
}

func (suite *KeeperTestSuite) TestConvertCoinNativeERC20() {
	var contractAddr common.Address

	testCases := []struct {
		name         string
		mint         int64
		convert      int64
		malleate     func(common.Address)
		extra        func()
		contractType int
		expPass      bool
	}{
		{
			"ok - sufficient funds",
			100,
			10,
			func(common.Address) {},
			func() {},
			contractMinterBurner,
			true,
		},
		{
			"ok - equal funds",
			100,
			100,
			func(common.Address) {},
			func() {},
			contractMinterBurner,
			true,
		},
		{
			"fail - insufficient funds",
			100,
			200,
			func(common.Address) {},
			func() {},
			contractMinterBurner,
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
		},
		{
			"fail - malicious delayed contract",
			100,
			10,
			func(common.Address) {},
			func() {},
			contractMaliciousDelayed,
			false,
		},
		{
			"fail - deleted module address - force fail",
			100,
			10,
			func(common.Address) {},
			func() {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, types.ModuleAddress.Bytes())
				suite.app.AccountKeeper.RemoveAccount(suite.ctx, acc)
			},
			contractMinterBurner,
			false,
		},
		{ //nolint:dupl
			"fail - force evm fail",
			100,
			10,
			func(common.Address) {},
			func() { //nolint:dupl
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			contractMinterBurner,
			false,
		},
		{
			"fail - force invalid transfer",
			100,
			10,
			func(common.Address) {},
			func() { //nolint:dupl
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			contractMinterBurner,
			false,
		},
		{
			"fail - force fail second balance",
			100,
			10,
			func(common.Address) {},
			func() { //nolint:dupl
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				balance[31] = uint8(1)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Twice()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("fail second balance"))
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			contractMinterBurner,
			false,
		},
		{
			"fail - force fail transfer",
			100,
			10,
			func(common.Address) {},
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			contractMinterBurner,
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			contractAddr = suite.setupRegisterERC20Pair(tc.contractType)
			suite.Require().NotNil(contractAddr)

			id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, contractAddr.String())
			pair, _ := suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
			coins := sdk.NewCoins(sdk.NewCoin(pair.Denom, sdk.NewInt(tc.mint)))
			coinName := types.CreateDenom(contractAddr.String())
			sender := sdk.AccAddress(suite.address.Bytes())

			// Precondition: Mint Coins to convert on sender account
			err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
			suite.Require().NoError(err)
			err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)
			suite.Require().NoError(err)
			cosmosBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sender, coinName)
			suite.Require().Equal(sdk.NewInt(tc.mint), cosmosBalance.Amount)

			// Precondition: Mint escrow tokens on module account
			suite.GrantERC20Token(contractAddr, suite.address, types.ModuleAddress, "MINTER_ROLE")
			suite.MintERC20Token(contractAddr, types.ModuleAddress, types.ModuleAddress, big.NewInt(tc.mint))
			tokenBalance := suite.BalanceOf(contractAddr, types.ModuleAddress)
			suite.Require().Equal(big.NewInt(tc.mint), tokenBalance)

			tc.malleate(contractAddr)
			suite.Commit()

			// Convert Coins back to ERC20s
			receiver := suite.address
			ctx := sdk.WrapSDKContext(suite.ctx)
			msg := types.NewMsgConvertCoin(
				sdk.NewCoin(coinName, sdk.NewInt(tc.convert)),
				receiver,
				sender,
			)

			tc.extra()
			res, err := suite.app.Erc20Keeper.ConvertCoin(ctx, msg)

			expRes := &types.MsgConvertCoinResponse{}
			suite.Commit()
			tokenBalance = suite.BalanceOf(contractAddr, suite.address)
			cosmosBalance = suite.app.BankKeeper.GetBalance(suite.ctx, sender, coinName)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expRes, res)
				suite.Require().Equal(sdk.NewInt(tc.mint-tc.convert), cosmosBalance.Amount)
				suite.Require().Equal(big.NewInt(tc.convert), tokenBalance.(*big.Int))
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
	suite.mintFeeCollector = false
}

func (suite *KeeperTestSuite) TestWrongPairOwnerERC20NativeCoin() {
	testCases := []struct {
		name      string
		mint      int64
		burn      int64
		reconvert int64
		expPass   bool
	}{
		{"ok - sufficient funds", 100, 10, 5, true},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			pair := suite.setupRegisterCoin(metadataCoin)
			suite.Require().NotNil(metadataCoin)
			suite.Require().NotNil(pair)

			// Precondition: Convert Coin to ERC20
			coins := sdk.NewCoins(sdk.NewCoin(cosmosTokenBase, sdk.NewInt(tc.mint)))
			sender := sdk.AccAddress(suite.address.Bytes())
			err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
			suite.Require().NoError(err)
			err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)
			suite.Require().NoError(err)
			msg := types.NewMsgConvertCoin(
				sdk.NewCoin(cosmosTokenBase, sdk.NewInt(tc.burn)),
				suite.address,
				sender,
			)

			pair.ContractOwner = types.OWNER_UNSPECIFIED
			suite.app.Erc20Keeper.SetTokenPair(suite.ctx, *pair)

			ctx := sdk.WrapSDKContext(suite.ctx)
			_, err = suite.app.Erc20Keeper.ConvertCoin(ctx, msg)
			suite.Require().Error(err, tc.name)

			// Convert ERC20s back to Coins
			ctx = sdk.WrapSDKContext(suite.ctx)
			contractAddr := common.HexToAddress(pair.Erc20Address)
			msgConvertERC20 := types.NewMsgConvertERC20(
				sdk.NewInt(tc.reconvert),
				sender,
				contractAddr,
				suite.address,
			)

			_, err = suite.app.Erc20Keeper.ConvertERC20(ctx, msgConvertERC20)
			suite.Require().Error(err, tc.name)
		})
	}
}

func (suite *KeeperTestSuite) TestConvertCoinNativeIBCVoucher() {
	testCases := []struct { //nolint:dupl
		name           string
		mint           int64
		burn           int64
		malleate       func(common.Address)
		extra          func()
		expPass        bool
		selfdestructed bool
	}{
		{
			"ok - sufficient funds",
			100,
			10,
			func(common.Address) {},
			func() {},
			true,
			false,
		},
		{
			"ok - equal funds",
			10,
			10,
			func(common.Address) {},
			func() {},
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
			true,
			true,
		},
		{
			"fail - insufficient funds",
			0,
			10,
			func(common.Address) {},
			func() {},
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
			false,
			false,
		},
		{
			"fail - deleted module account - force fail", 100, 10, func(common.Address) {},
			func() {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, types.ModuleAddress.Bytes())
				suite.app.AccountKeeper.RemoveAccount(suite.ctx, acc)
			}, false, false,
		},
		{ //nolint:dupl
			"fail - force evm fail", 100, 10, func(common.Address) {},
			func() { //nolint:dupl
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			}, false, false,
		},
		{
			"fail - force evm balance error", 100, 10, func(common.Address) {},
			func() { //nolint:dupl
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				// first balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				// convert coin
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil).Once()
				// second balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, fmt.Errorf("third")).Once()
				// Extra call on test
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			}, false, false,
		},
		{
			"fail - force balance error", 100, 10, func(common.Address) {},
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Times(4)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			}, false, false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			pair := suite.setupRegisterCoin(metadataIbc)
			suite.Require().NotNil(metadataIbc)
			erc20 := pair.GetERC20Contract()
			tc.malleate(erc20)
			suite.Commit()

			ctx := sdk.WrapSDKContext(suite.ctx)
			coins := sdk.NewCoins(sdk.NewCoin(ibcBase, sdk.NewInt(tc.mint)))
			sender := sdk.AccAddress(suite.address.Bytes())
			msg := types.NewMsgConvertCoin(
				sdk.NewCoin(ibcBase, sdk.NewInt(tc.burn)),
				suite.address,
				sender,
			)

			suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins) //nolint:errcheck
			suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)

			tc.extra()
			res, err := suite.app.Erc20Keeper.ConvertCoin(ctx, msg)
			expRes := &types.MsgConvertCoinResponse{}
			suite.Commit()
			balance := suite.BalanceOf(common.HexToAddress(pair.Erc20Address), suite.address)
			cosmosBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sender, metadataIbc.Base)

			if tc.expPass {
				suite.Require().NoError(err, tc.name)

				acc := suite.app.EvmKeeper.GetAccountWithoutBalance(suite.ctx, erc20)
				if tc.selfdestructed {
					suite.Require().Nil(acc, "expected contract to be destroyed")
				} else {
					suite.Require().NotNil(acc)
				}

				if tc.selfdestructed || !acc.IsContract() {
					id := suite.app.Erc20Keeper.GetTokenPairID(suite.ctx, erc20.String())
					_, found := suite.app.Erc20Keeper.GetTokenPair(suite.ctx, id)
					suite.Require().False(found)
				} else {
					suite.Require().Equal(expRes, res)
					suite.Require().Equal(cosmosBalance.Amount.Int64(), sdk.NewInt(tc.mint-tc.burn).Int64())
					suite.Require().Equal(balance.(*big.Int).Int64(), big.NewInt(tc.burn).Int64())
				}
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
	suite.mintFeeCollector = false
}

func (suite *KeeperTestSuite) TestConvertERC20NativeIBCVoucher() {
	testCases := []struct {
		name      string
		mint      int64
		burn      int64
		reconvert int64
		malleate  func()
		expPass   bool
	}{
		{"ok - sufficient funds", 100, 10, 5, func() {}, true},
		{"ok - equal funds", 10, 10, 10, func() {}, true},
		{"fail - insufficient funds", 10, 1, 5, func() {}, false},
		{"fail ", 10, 1, -5, func() {}, false},
		{
			"fail - deleted module account - force fail", 100, 10, 5,
			func() {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, types.ModuleAddress.Bytes())
				suite.app.AccountKeeper.RemoveAccount(suite.ctx, acc)
			},
			false,
		},
		{ //nolint:dupl
			"fail - force evm fail", 100, 10, 5,
			func() { //nolint:dupl
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			false,
		},
		{
			"fail - force fail second balance", 100, 10, 5,
			func() { //nolint:dupl
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				// first balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				// convert coin
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil).Once()
				// second balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, fmt.Errorf("third")).Once()
				// Extra call on test
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			false,
		},
		{
			"fail - force fail second balance", 100, 10, 5,
			func() {
				mockEVMKeeper := &MockEVMKeeper{}
				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				existingAcc := &statedb.Account{Nonce: uint64(1), Balance: common.Big1}
				balance := make([]uint8, 32)
				mockEVMKeeper.On("EstimateGas", mock.Anything, mock.Anything).Return(&evmtypes.EstimateGasResponse{Gas: uint64(200)}, nil)
				// first balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				// convert coin
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil).Once()
				// second balance of
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
				// Extra call on test
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{}, nil)
				mockEVMKeeper.On("GetAccountWithoutBalance", mock.Anything, mock.Anything).Return(existingAcc, nil)
			},
			false,
		},
		{
			"fail - force fail unescrow", 100, 10, 5,
			func() {
				mockBankKeeper := &MockBankKeeper{}

				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					mockBankKeeper, suite.app.EvmKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				mockBankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to unescrow"))
				mockBankKeeper.On("BlockedAddr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)
				mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sdk.Coin{Denom: "coin", Amount: sdk.OneInt()})
			},
			false,
		},
		{
			"fail - force fail balance after transfer", 100, 10, 5,
			func() {
				mockBankKeeper := &MockBankKeeper{}

				suite.app.Erc20Keeper = keeper.NewKeeper(
					suite.app.GetKey("erc20"), suite.app.AppCodec(),
					authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
					mockBankKeeper, suite.app.EvmKeeper, suite.app.StakingKeeper, suite.app.ClaimsKeeper)

				mockBankKeeper.On("SendCoinsFromModuleToAccount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockBankKeeper.On("BlockedAddr", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false)
				mockBankKeeper.On("GetBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(sdk.Coin{Denom: ibcBase, Amount: sdk.OneInt()})
			},
			false,
		},
	}
	for _, tc := range testCases { //nolint:dupl
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			pair := suite.setupRegisterCoin(metadataIbc)
			suite.Require().NotNil(metadataIbc)
			suite.Require().NotNil(pair)

			// Precondition: Convert Coin to ERC20
			coins := sdk.NewCoins(sdk.NewCoin(ibcBase, sdk.NewInt(tc.mint)))
			sender := sdk.AccAddress(suite.address.Bytes())
			err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
			suite.Require().NoError(err, tc.name)
			err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)
			suite.Require().NoError(err, tc.name)
			msg := types.NewMsgConvertCoin(
				sdk.NewCoin(ibcBase, sdk.NewInt(tc.burn)),
				suite.address,
				sender,
			)

			ctx := sdk.WrapSDKContext(suite.ctx)
			_, err = suite.app.Erc20Keeper.ConvertCoin(ctx, msg)
			suite.Require().NoError(err, tc.name)
			suite.Commit()
			balance := suite.BalanceOf(common.HexToAddress(pair.Erc20Address), suite.address)
			cosmosBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sender, metadataIbc.Base)
			suite.Require().Equal(cosmosBalance.Amount.Int64(), sdk.NewInt(tc.mint-tc.burn).Int64())
			suite.Require().Equal(balance, big.NewInt(tc.burn))

			// Convert ERC20s back to Coins
			ctx = sdk.WrapSDKContext(suite.ctx)
			contractAddr := common.HexToAddress(pair.Erc20Address)
			msgConvertERC20 := types.NewMsgConvertERC20(
				sdk.NewInt(tc.reconvert),
				sender,
				contractAddr,
				suite.address,
			)

			tc.malleate()
			res, err := suite.app.Erc20Keeper.ConvertERC20(ctx, msgConvertERC20)
			expRes := &types.MsgConvertERC20Response{}
			suite.Commit()
			balance = suite.BalanceOf(contractAddr, suite.address)
			cosmosBalance = suite.app.BankKeeper.GetBalance(suite.ctx, sender, pair.Denom)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expRes, res)
				suite.Require().Equal(cosmosBalance.Amount.Int64(), sdk.NewInt(tc.mint-tc.burn+tc.reconvert).Int64())
				suite.Require().Equal(balance.(*big.Int).Int64(), big.NewInt(tc.burn-tc.reconvert).Int64())
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
