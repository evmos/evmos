package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/tharsis/ethermint/tests"

	"github.com/tharsis/evmos/x/claims/types"
	inflationtypes "github.com/tharsis/evmos/x/inflation/types"
)

func (suite *KeeperTestSuite) TestEndBlock() {
	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"claim disabled",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = false
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
			},
		},
		{
			"not claim time",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = true
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
			},
		},
		{
			"claim enabled",
			func() {
				params := suite.app.ClaimsKeeper.GetParams(suite.ctx)
				params.EnableClaims = true
				params.AirdropStartTime = time.Time{}
				params.DurationUntilDecay = time.Hour
				params.DurationOfDecay = time.Hour
				suite.app.ClaimsKeeper.SetParams(suite.ctx, params)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			suite.app.ClaimsKeeper.EndBlocker(suite.ctx)
		})
	}
}

func (suite *KeeperTestSuite) TestClawbackEmptyAccounts() {
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())

	var amount int64 = 10000

	testCases := []struct {
		name     string
		funds    int64
		malleate func()
	}{
		{
			"no claims records",
			0,
			func() {
			},
		},
		{
			"no account",
			0,
			func() {
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})
			},
		},
		{
			"sequence not zero",
			0,
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 1))
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})
			},
		},
		{
			"no balance",
			0,
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})
			},
		},
		{
			"balance non zero",
			amount,
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))

				coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(amount)))
				err := suite.app.BankKeeper.MintCoins(suite.ctx, inflationtypes.ModuleName, coins)
				suite.Require().NoError(err)
				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, inflationtypes.ModuleName, addr, coins)
				suite.Require().NoError(err)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})
			},
		},
		{
			"balance non zero not claim denom",
			0,
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))

				coins := sdk.NewCoins(sdk.NewCoin("testcoin", sdk.NewInt(amount)))
				err := suite.app.BankKeeper.MintCoins(suite.ctx, inflationtypes.ModuleName, coins)
				suite.Require().NoError(err)
				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, inflationtypes.ModuleName, addr, coins)
				suite.Require().NoError(err)
				suite.app.ClaimsKeeper.SetClaimsRecord(suite.ctx, addr, types.ClaimsRecord{})
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			suite.app.ClaimsKeeper.ClawbackEmptyAccounts(suite.ctx, "aevmos")

			acc := suite.app.AccountKeeper.GetModuleAccount(suite.ctx, distrtypes.ModuleName)
			balance := suite.app.BankKeeper.GetBalance(suite.ctx, acc.GetAddress(), "aevmos")
			suite.Require().Equal(balance.Amount, sdk.NewInt(tc.funds))
		})
	}
}

func (suite *KeeperTestSuite) TestClawbackEscrowedTokensABCI() {
	var amount int64 = 10000

	testCases := []struct {
		name     string
		funds    int64
		malleate func()
	}{
		{
			"no balance",
			0,
			func() {
			},
		},
		{
			"balance on module account",
			amount,
			func() {
				coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(amount)))
				err := suite.app.BankKeeper.MintCoins(suite.ctx, inflationtypes.ModuleName, coins)
				suite.Require().NoError(err)
				err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, inflationtypes.ModuleName, types.ModuleName, coins)
				suite.Require().NoError(err)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			err := suite.app.ClaimsKeeper.ClawbackEscrowedTokens(suite.ctx)
			suite.Require().NoError(err)

			acc := suite.app.AccountKeeper.GetModuleAccount(suite.ctx, distrtypes.ModuleName)
			balance := suite.app.BankKeeper.GetBalance(suite.ctx, acc.GetAddress(), "aevmos")
			suite.Require().Equal(balance.Amount, sdk.NewInt(tc.funds))
		})
	}
}
