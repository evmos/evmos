package keeper_test

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

func (suite *KeeperTestSuite) TestConvertCoinNativeCoin() {
	testCases := []struct {
		name     string
		mint     int64
		burn     int64
		malleate func()
		expPass  bool
	}{
		{"ok - sufficient funds", 100, 10, func() {}, true},
		{"ok - equal funds", 10, 10, func() {}, true},
		{"fail - insufficient funds", 0, 10, func() {}, false},
		{
			"fail - minting disabled",
			100,
			10,
			func() {
				params := types.DefaultParams()
				params.EnableIntrarelayer = false
				suite.app.IntrarelayerKeeper.SetParams(suite.ctx, params)
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			metadata, pair := suite.setupRegisterCoin()
			suite.Require().NotNil(metadata)

			tc.malleate()
			suite.Commit()

			ctx := sdk.WrapSDKContext(suite.ctx)
			coins := sdk.NewCoins(sdk.NewCoin(cosmosTokenName, sdk.NewInt(tc.mint)))
			sender := sdk.AccAddress(suite.address.Bytes())
			msg := types.NewMsgConvertCoin(
				sdk.NewCoin(cosmosTokenName, sdk.NewInt(tc.burn)),
				suite.address,
				sender,
			)

			suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
			suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)

			res, err := suite.app.IntrarelayerKeeper.ConvertCoin(ctx, msg)
			expRes := &types.MsgConvertCoinResponse{}
			suite.Commit()
			balance := suite.BalanceOf(common.HexToAddress(pair.Erc20Address), suite.address)
			cosmosBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sender, metadata.Base)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expRes, res)
				suite.Require().Equal(cosmosBalance.Amount.Int64(), sdk.NewInt(tc.mint-tc.burn).Int64())
				suite.Require().Equal(balance.(*big.Int).Int64(), big.NewInt(tc.burn).Int64())

			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
	suite.mintFeeCollector = false
}

func (suite *KeeperTestSuite) TestConvertECR20NativeCoin() {
	testCases := []struct {
		name      string
		mint      int64
		burn      int64
		reconvert int64
		expPass   bool
	}{
		{"ok - sufficient funds", 100, 10, 5, true},
		{"ok - equal funds", 10, 10, 10, true},
		{"fail - insufficient funds", 10, 1, 5, false},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			metadata, pair := suite.setupRegisterCoin()
			suite.Require().NotNil(metadata)
			suite.Require().NotNil(pair)

			// Precondition: Convert Coin to ERC20
			coins := sdk.NewCoins(sdk.NewCoin(cosmosTokenName, sdk.NewInt(tc.mint)))
			sender := sdk.AccAddress(suite.address.Bytes())
			suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
			suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)
			msg := types.NewMsgConvertCoin(
				sdk.NewCoin(cosmosTokenName, sdk.NewInt(tc.burn)),
				suite.address,
				sender,
			)

			ctx := sdk.WrapSDKContext(suite.ctx)
			_, err := suite.app.IntrarelayerKeeper.ConvertCoin(ctx, msg)
			suite.Require().NoError(err, tc.name)
			suite.Commit()
			balance := suite.BalanceOf(common.HexToAddress(pair.Erc20Address), suite.address)
			cosmosBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sender, metadata.Base)
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

			res, err := suite.app.IntrarelayerKeeper.ConvertERC20(ctx, msgConvertERC20)
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

func (suite *KeeperTestSuite) TestConvertECR20NativeERC20() {
	testCases := []struct {
		name     string
		mint     int64
		burn     int64
		malleate func()
		expPass  bool
	}{
		{"ok - sufficient funds", 100, 10, func() {}, true},
		{"ok - equal funds", 10, 10, func() {}, true},
		{"fail - insufficient funds - callEVM", 0, 10, func() {}, false},
		{
			"fail - minting disabled",
			100,
			10,
			func() {
				params := types.DefaultParams()
				params.EnableIntrarelayer = false
				suite.app.IntrarelayerKeeper.SetParams(suite.ctx, params)
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			contractAddr := suite.setupRegisterERC20Pair()
			suite.Require().NotNil(contractAddr)

			tc.malleate()
			suite.Commit()

			coinName := "irm" + contractAddr.String()
			sender := sdk.AccAddress(suite.address.Bytes())
			msg := types.NewMsgConvertERC20(
				sdk.NewInt(tc.burn),
				sender,
				contractAddr,
				suite.address,
			)
			suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(tc.mint))
			suite.Commit()
			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.app.IntrarelayerKeeper.ConvertERC20(ctx, msg)
			expRes := &types.MsgConvertERC20Response{}
			suite.Commit()
			balance := suite.BalanceOf(contractAddr, suite.address)
			cosmosBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sender, coinName)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expRes, res)
				suite.Require().Equal(cosmosBalance.Amount, sdk.NewInt(tc.burn))
				suite.Require().Equal(balance.(*big.Int).Int64(), big.NewInt(tc.mint-tc.burn).Int64())

			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
	suite.mintFeeCollector = false
}

func (suite *KeeperTestSuite) TestConvertCoinNativeERC20() {
	testCases := []struct {
		name      string
		mint      int64
		burn      int64
		reconvert int64
		expPass   bool
	}{
		{"ok - sufficient funds", 100, 10, 5, true},
		{"ok - equal funds", 10, 10, 10, true},
		{"fail - insufficient funds", 10, 1, 5, false},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.mintFeeCollector = true
			suite.SetupTest()
			contractAddr := suite.setupRegisterERC20Pair()
			suite.Require().NotNil(contractAddr)

			// Precondition: Convert ERC20 to Coins
			coinName := "irm" + contractAddr.String()
			sender := sdk.AccAddress(suite.address.Bytes())
			suite.MintERC20Token(contractAddr, suite.address, suite.address, big.NewInt(tc.mint))
			suite.Commit()
			msgConvertERC20 := types.NewMsgConvertERC20(
				sdk.NewInt(tc.burn),
				sender,
				contractAddr,
				suite.address,
			)

			ctx := sdk.WrapSDKContext(suite.ctx)
			_, err := suite.app.IntrarelayerKeeper.ConvertERC20(ctx, msgConvertERC20)
			suite.Require().NoError(err)
			suite.Commit()
			balance := suite.BalanceOf(contractAddr, suite.address)
			cosmosBalance := suite.app.BankKeeper.GetBalance(suite.ctx, sender, coinName)
			suite.Require().Equal(cosmosBalance.Amount, sdk.NewInt(tc.burn))
			suite.Require().Equal(balance.(*big.Int).Int64(), big.NewInt(tc.mint-tc.burn).Int64())

			// Convert Coins back to ERC20s
			ctx = sdk.WrapSDKContext(suite.ctx)
			msg := types.NewMsgConvertCoin(
				sdk.NewCoin(coinName, sdk.NewInt(tc.reconvert)),
				suite.address,
				sender,
			)

			res, err := suite.app.IntrarelayerKeeper.ConvertCoin(ctx, msg)
			expRes := &types.MsgConvertCoinResponse{}
			suite.Commit()
			balance = suite.BalanceOf(contractAddr, suite.address)
			cosmosBalance = suite.app.BankKeeper.GetBalance(suite.ctx, sender, coinName)
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expRes, res)
				suite.Require().Equal(cosmosBalance.Amount.Int64(), sdk.NewInt(tc.burn-tc.reconvert).Int64())
				suite.Require().Equal(balance.(*big.Int).Int64(), big.NewInt(tc.mint-tc.burn+tc.reconvert).Int64())

			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
	suite.mintFeeCollector = false
}
