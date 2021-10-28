package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

func (suite *KeeperTestSuite) TestConvertCoin() {
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		// {"coin not registered", func() {}, false},
		// {
		// 	"coin registered - insufficient funds",
		// 	func() {
		// 		pair := types.NewTokenPair(tests.GenerateAddress(), erc20Name, true)
		// 		id := pair.GetID()
		// 		suite.app.IntrarelayerKeeper.SetTokenPair(suite.ctx, pair)
		// 		suite.app.IntrarelayerKeeper.SetDenomMap(suite.ctx, pair.Denom, id)
		// 		suite.app.IntrarelayerKeeper.SetERC20Map(suite.ctx, pair.GetERC20Contract(), id)
		// 	},
		// 	false,
		// },
		{

			"ok - coin registered - sufficient funds - callEVM",
			func() {
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			contractAddr := suite.setupNewTokenPair()
			suite.Require().NotNil(contractAddr)
			// id := suite.app.IntrarelayerKeeper.GetTokenPairID(suite.ctx, contractAddr.String())
			// pair, _ := suite.app.IntrarelayerKeeper.GetTokenPair(suite.ctx, id)

			sender := sdk.AccAddress(tests.GenerateAddress().Bytes())
			receiver := tests.GenerateAddress()
			coins := sdk.NewCoins(sdk.NewCoin(cosmosTokenName, sdk.NewInt(100)))
			suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
			suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sender, coins)

			ctx := sdk.WrapSDKContext(suite.ctx)
			msg := types.NewMsgConvertCoin(
				sdk.NewCoin(cosmosTokenName, sdk.NewInt(100)),
				receiver,
				sender,
			)
			res, err := suite.app.IntrarelayerKeeper.ConvertCoin(ctx, msg)
			expRes := &types.MsgConvertCoinResponse{}
			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestConvertECR20() {
	erc20 := tests.GenerateAddress()
	// denom := "coin"
	// pair := types.NewTokenPair(erc20, denom, true)
	// id := pair.GetID()

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{"coin not registered", func() {}, false},
		// TODO: use burn contract with ABI
		// {
		// 	"erc20 has no burn method",
		// 	func() {
		// 		suite.app.IntrarelayerKeeper.SetTokenPair(suite.ctx, pair)
		// 		suite.app.IntrarelayerKeeper.SetDenomMap(suite.ctx, pair.Denom, id)
		// 		suite.app.IntrarelayerKeeper.SetERC20Map(suite.ctx, pair.GetERC20Contract(), id)
		// 	},
		// 	true,
		// },
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			msg := types.NewMsgConvertERC20(
				sdk.NewInt(100),
				sdk.AccAddress{},
				erc20,
				tests.GenerateAddress(),
			)
			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.app.IntrarelayerKeeper.ConvertERC20(ctx, msg)
			expRes := &types.MsgConvertERC20Response{}

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err, tc.name)
			}
		})
	}
}
