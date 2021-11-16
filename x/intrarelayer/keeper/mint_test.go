package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/x/intrarelayer/types"
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
			"intrarelaying is disabled globally",
			func() {
				params := types.DefaultParams()
				params.EnableIntrarelayer = false
				suite.app.IntrarelayerKeeper.SetParams(suite.ctx, params)
			},
			false,
		},
		{
			"token pair not found",
			func() {},
			false,
		},
		{
			"intrarelaying is disabled for the given pair",
			func() {
				expPair.Enabled = false
				suite.app.IntrarelayerKeeper.SetTokenPair(suite.ctx, expPair)
				suite.app.IntrarelayerKeeper.SetDenomMap(suite.ctx, expPair.Denom, id)
				suite.app.IntrarelayerKeeper.SetERC20Map(suite.ctx, expPair.GetERC20Contract(), id)
			},
			false,
		},
		{
			"token transfers are disabled",
			func() {
				expPair.Enabled = true
				suite.app.IntrarelayerKeeper.SetTokenPair(suite.ctx, expPair)
				suite.app.IntrarelayerKeeper.SetDenomMap(suite.ctx, expPair.Denom, id)
				suite.app.IntrarelayerKeeper.SetERC20Map(suite.ctx, expPair.GetERC20Contract(), id)

				params := banktypes.DefaultParams()
				params.SendEnabled = []*banktypes.SendEnabled{
					{Denom: expPair.Denom, Enabled: false},
				}
				suite.app.BankKeeper.SetParams(suite.ctx, params)
			},
			false,
		},
		{
			"ok",
			func() {
				suite.app.IntrarelayerKeeper.SetTokenPair(suite.ctx, expPair)
				suite.app.IntrarelayerKeeper.SetDenomMap(suite.ctx, expPair.Denom, id)
				suite.app.IntrarelayerKeeper.SetERC20Map(suite.ctx, expPair.GetERC20Contract(), id)
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			pair, err := suite.app.IntrarelayerKeeper.MintingEnabled(suite.ctx, sender, receiver, expPair.Erc20Address)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expPair, pair)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
