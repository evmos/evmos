package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/ethermint/tests"

	"github.com/evmos/evmos/v7/x/erc20/types"
)

func (suite *KeeperTestSuite) TestMintingEnabled() {
	sender := sdk.AccAddress(tests.GenerateAddress().Bytes())
	receiver := sdk.AccAddress(tests.GenerateAddress().Bytes())
	expPair := types.NewTokenPair(tests.GenerateAddress(), "coin", true, types.OWNER_MODULE)
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
				suite.app.Erc20Keeper.SetParams(suite.ctx, params)
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
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, expPair)
				suite.app.Erc20Keeper.SetDenomMap(suite.ctx, expPair.Denom, id)
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, expPair.GetERC20Contract(), id)
			},
			false,
		},
		{
			"token transfers are disabled",
			func() {
				expPair.Enabled = true
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, expPair)
				suite.app.Erc20Keeper.SetDenomMap(suite.ctx, expPair.Denom, id)
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, expPair.GetERC20Contract(), id)

				params := banktypes.DefaultParams()
				params.SendEnabled = []*banktypes.SendEnabled{
					{Denom: expPair.Denom, Enabled: false},
				}
				suite.app.BankKeeper.SetParams(suite.ctx, params)
			},
			false,
		},
		{
			"token not registered",
			func() {
				suite.app.Erc20Keeper.SetDenomMap(suite.ctx, expPair.Denom, id)
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, expPair.GetERC20Contract(), id)
			},
			false,
		},
		{
			"receiver address is blocked (module account)",
			func() {
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, expPair)
				suite.app.Erc20Keeper.SetDenomMap(suite.ctx, expPair.Denom, id)
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, expPair.GetERC20Contract(), id)

				acc := suite.app.AccountKeeper.GetModuleAccount(suite.ctx, types.ModuleName)
				receiver = acc.GetAddress()
			},
			false,
		},
		{
			"ok",
			func() {
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, expPair)
				suite.app.Erc20Keeper.SetDenomMap(suite.ctx, expPair.Denom, id)
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, expPair.GetERC20Contract(), id)

				receiver = sdk.AccAddress(tests.GenerateAddress().Bytes())
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			pair, err := suite.app.Erc20Keeper.MintingEnabled(suite.ctx, sender, receiver, expPair.Erc20Address)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expPair, pair)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
