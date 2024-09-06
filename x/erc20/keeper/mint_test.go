package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	utiltx "github.com/evmos/evmos/v19/testutil/tx"
	"github.com/evmos/evmos/v19/x/erc20/types"
)

func (suite *KeeperTestSuite) TestMintingEnabled() {
	var ctx sdk.Context
	sender := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	receiver := sdk.AccAddress(utiltx.GenerateAddress().Bytes())
	expPair := types.NewTokenPair(utiltx.GenerateAddress(), "coin", types.OWNER_MODULE)
	id := expPair.GetID()

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"conversion is disabled globally",
			func() {
				params := types.DefaultParams()
				params.EnableErc20 = false
				suite.network.App.Erc20Keeper.SetParams(ctx, params) //nolint:errcheck
			},
			false,
		},
		{
			"token pair not found",
			func() {},
			false,
		},
		{
			"conversion is disabled for the given pair",
			func() {
				expPair.Enabled = false
				suite.network.App.Erc20Keeper.SetTokenPair(ctx, expPair)
				suite.network.App.Erc20Keeper.SetDenomMap(ctx, expPair.Denom, id)
				suite.network.App.Erc20Keeper.SetERC20Map(ctx, expPair.GetERC20Contract(), id)
			},
			false,
		},
		{
			"token transfers are disabled",
			func() {
				expPair.Enabled = true
				suite.network.App.Erc20Keeper.SetTokenPair(ctx, expPair)
				suite.network.App.Erc20Keeper.SetDenomMap(ctx, expPair.Denom, id)
				suite.network.App.Erc20Keeper.SetERC20Map(ctx, expPair.GetERC20Contract(), id)

				params := banktypes.DefaultParams()
				params.SendEnabled = []*banktypes.SendEnabled{ //nolint:staticcheck
					{Denom: expPair.Denom, Enabled: false},
				}
				err := suite.network.App.BankKeeper.SetParams(ctx, params)
				suite.Require().NoError(err)
			},
			false,
		},
		{
			"token not registered",
			func() {
				suite.network.App.Erc20Keeper.SetDenomMap(ctx, expPair.Denom, id)
				suite.network.App.Erc20Keeper.SetERC20Map(ctx, expPair.GetERC20Contract(), id)
			},
			false,
		},
		{
			"receiver address is blocked (module account)",
			func() {
				suite.network.App.Erc20Keeper.SetTokenPair(ctx, expPair)
				suite.network.App.Erc20Keeper.SetDenomMap(ctx, expPair.Denom, id)
				suite.network.App.Erc20Keeper.SetERC20Map(ctx, expPair.GetERC20Contract(), id)

				acc := suite.network.App.AccountKeeper.GetModuleAccount(ctx, types.ModuleName)
				receiver = acc.GetAddress()
			},
			false,
		},
		{
			"ok",
			func() {
				suite.network.App.Erc20Keeper.SetTokenPair(ctx, expPair)
				suite.network.App.Erc20Keeper.SetDenomMap(ctx, expPair.Denom, id)
				suite.network.App.Erc20Keeper.SetERC20Map(ctx, expPair.GetERC20Contract(), id)

				receiver = sdk.AccAddress(utiltx.GenerateAddress().Bytes())
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			ctx = suite.network.GetContext()

			tc.malleate()

			pair, err := suite.network.App.Erc20Keeper.MintingEnabled(ctx, sender, receiver, expPair.Erc20Address)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expPair, pair)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
