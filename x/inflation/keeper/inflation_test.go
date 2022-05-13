package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	incentivestypes "github.com/tharsis/evmos/v4/x/incentives/types"
	"github.com/tharsis/evmos/v4/x/inflation/types"
)

func (suite *KeeperTestSuite) TestMintAndAllocateInflation() {
	testCases := []struct {
		name                  string
		mintCoin              sdk.Coin
		malleate              func()
		expStakingRewardAmt   sdk.Coin
		expUsageIncentivesAmt sdk.Coin
		expCommunityPoolAmt   sdk.DecCoins
		expPass               bool
	}{
		{
			"pass",
			sdk.NewCoin(denomMint, sdk.NewInt(1_000_000)),
			func() {},
			sdk.NewCoin(denomMint, sdk.NewInt(533_333)),
			sdk.NewCoin(denomMint, sdk.NewInt(333_333)),
			sdk.NewDecCoins(sdk.NewDecCoin(denomMint, sdk.NewInt(133_334))),
			true,
		},
		{
			"pass - no coins minted ",
			sdk.NewCoin(denomMint, sdk.ZeroInt()),
			func() {},
			sdk.NewCoin(denomMint, sdk.ZeroInt()),
			sdk.NewCoin(denomMint, sdk.ZeroInt()),
			sdk.DecCoins(nil),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			err := suite.app.InflationKeeper.MintAndAllocateInflation(suite.ctx, tc.mintCoin)

			// Get balances
			balanceModule := suite.app.BankKeeper.GetBalance(
				suite.ctx,
				suite.app.AccountKeeper.GetModuleAddress(types.ModuleName),
				denomMint,
			)

			feeCollector := suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
			balanceStakingRewards := suite.app.BankKeeper.GetBalance(
				suite.ctx,
				feeCollector,
				denomMint,
			)

			incentives := suite.app.AccountKeeper.GetModuleAddress(incentivestypes.ModuleName)
			balanceUsageIncentives := suite.app.BankKeeper.GetBalance(
				suite.ctx,
				incentives,
				denomMint,
			)

			balanceCommunityPool := suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().True(balanceModule.IsZero())
				suite.Require().Equal(tc.expStakingRewardAmt, balanceStakingRewards)
				suite.Require().Equal(tc.expUsageIncentivesAmt, balanceUsageIncentives)
				suite.Require().Equal(tc.expCommunityPoolAmt, balanceCommunityPool)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetCirculatingSupplyAndInflationRate() {
	testCases := []struct {
		name             string
		bankSupply       int64
		malleate         func()
		expInflationRate sdk.Dec
	}{
		{
			"no mint provision",
			400_000_000,
			func() {
				suite.app.InflationKeeper.SetEpochMintProvision(suite.ctx, sdk.ZeroDec())
			},
			sdk.ZeroDec(),
		},
		{
			"no epochs per period",
			400_000_000,
			func() {
				suite.app.InflationKeeper.SetEpochsPerPeriod(suite.ctx, 0)
			},
			sdk.ZeroDec(),
		},
		{
			"high supply",
			800_000_000,
			func() {},
			sdk.MustNewDecFromStr("51.562500000000000000"),
		},
		{
			"low supply",
			400_000_000,
			func() {},
			sdk.MustNewDecFromStr("154.687500000000000000"),
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			// Team allocation is only set on mainnet
			suite.ctx = suite.ctx.WithChainID("evmos_9001-1")
			tc.malleate()

			// Mint coins to increase supply
			coin := sdk.NewCoin(types.DefaultInflationDenom, sdk.TokensFromConsensusPower(tc.bankSupply, sdk.DefaultPowerReduction))
			decCoin := sdk.NewDecCoinFromCoin(coin)
			suite.app.InflationKeeper.MintCoins(suite.ctx, coin)

			teamAlloc := sdk.NewDecCoin(types.DefaultInflationDenom, sdk.TokensFromConsensusPower(int64(200_000_000), sdk.DefaultPowerReduction))
			circulatingSupply := s.app.InflationKeeper.GetCirculatingSupply(suite.ctx)

			suite.Require().Equal(decCoin.Sub(teamAlloc).Amount, circulatingSupply)

			inflationRate := s.app.InflationKeeper.GetInflationRate(suite.ctx)
			suite.Require().Equal(tc.expInflationRate, inflationRate)
		})
	}
}
