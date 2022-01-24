package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"github.com/tharsis/evmos/x/claim/types"
)

func (suite *KeeperTestSuite) TestEndBlock() {

	testCases := []struct {
		name      string
		NoBaseFee bool
		malleate  func()
	}{
		{
			"claim enabled ",
			true,
			func() {
				params := suite.app.ClaimKeeper.GetParams(suite.ctx)
				params.EnableClaim = true
				params.AirdropStartTime = time.Time{}
				params.DurationUntilDecay = time.Hour
				params.DurationOfDecay = time.Hour
				suite.app.ClaimKeeper.SetParams(suite.ctx, params)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			suite.app.ClaimKeeper.EndBlocker(suite.ctx)
		})
	}
}

func (suite *KeeperTestSuite) TestClawbackEmptyAccounts() {

	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	addr := sdk.AccAddress(priv.PubKey().Address())

	testCases := []struct {
		name      string
		NoBaseFee bool
		malleate  func()
	}{
		{
			"No claims records ",
			true,
			func() {

			},
		},
		{
			"No account ",
			true,
			func() {
				suite.app.ClaimKeeper.SetClaimRecord(suite.ctx, addr, types.ClaimRecord{})
			},
		},
		{
			"Sequence not zero ",
			true,
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 1))
				suite.app.ClaimKeeper.SetClaimRecord(suite.ctx, addr, types.ClaimRecord{})
			},
		},
		{
			"No balance ",
			true,
			func() {
				suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))
				suite.app.ClaimKeeper.SetClaimRecord(suite.ctx, addr, types.ClaimRecord{})
			},
		},
		{
			"Balance non zero ",
			true,
			func() {

				suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))

				coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(10400)))
				_ = suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins)
				_ = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, addr, coins)

				suite.app.ClaimKeeper.SetClaimRecord(suite.ctx, addr, types.ClaimRecord{})
			},
		},
		{
			"Balance non zero not claim denom  ",
			true,
			func() {

				suite.app.AccountKeeper.SetAccount(suite.ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))

				coins := sdk.NewCoins(sdk.NewCoin("testcoin", sdk.NewInt(10400)))
				_ = suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins)
				_ = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, minttypes.ModuleName, addr, coins)

				suite.app.ClaimKeeper.SetClaimRecord(suite.ctx, addr, types.ClaimRecord{})
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			suite.app.ClaimKeeper.ClawbackEmptyAccounts(suite.ctx, "aevmos")

		})
	}
}

func (suite *KeeperTestSuite) TestClawbackEscrowedTokensABCI() {

	testCases := []struct {
		name      string
		NoBaseFee bool
		malleate  func()
	}{
		{
			"No balance",
			true,
			func() {

			},
		},
		{
			"Balance on module account",
			true,
			func() {
				coins := sdk.NewCoins(sdk.NewCoin("aevmos", sdk.NewInt(10400)))
				_ = suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins)
				_ = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, minttypes.ModuleName, types.ModuleName, coins)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			suite.app.ClaimKeeper.ClawbackEscrowedTokens(suite.ctx)

		})
	}
}
