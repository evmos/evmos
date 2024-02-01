package keeper_test

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	evmostypes "github.com/evmos/evmos/v16/types"
	"github.com/evmos/evmos/v16/x/inflation/v1/types"
)

func (suite *KeeperTestSuite) TestMintAndAllocateInflation() {
	testCases := []struct {
		name                string
		mintCoin            sdk.Coin
		malleate            func()
		expStakingRewardAmt sdk.Coin
		expCommunityPoolAmt sdk.DecCoins
		expPass             bool
	}{
		{
			"pass",
			sdk.NewCoin(denomMint, math.NewInt(1_000_000)),
			func() {},
			sdk.NewCoin(denomMint, math.NewInt(533_333)),
			sdk.NewDecCoins(sdk.NewDecCoin(denomMint, math.NewInt(466_667))),
			true,
		},
		{
			"pass - no coins minted ",
			sdk.NewCoin(denomMint, math.ZeroInt()),
			func() {},
			sdk.NewCoin(denomMint, math.ZeroInt()),
			sdk.DecCoins(nil),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			_, _, err := suite.app.InflationKeeper.MintAndAllocateInflation(suite.ctx, tc.mintCoin, types.DefaultParams())

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

			balanceCommunityPool := suite.app.DistrKeeper.GetFeePoolCommunityCoins(suite.ctx)

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().True(balanceModule.IsZero())
				suite.Require().Equal(tc.expStakingRewardAmt, balanceStakingRewards)
				suite.Require().Equal(tc.expCommunityPoolAmt, balanceCommunityPool)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetCirculatingSupplyAndInflationRate() {
	// the total bonded tokens for the 2 accounts initialized on the setup
	bondedAmt := math.NewInt(1000100000000000000)
	bondedCoins := sdk.NewDecCoin(evmostypes.AttoEvmos, bondedAmt)

	testCases := []struct {
		name             string
		bankSupply       math.Int
		malleate         func()
		expInflationRate math.LegacyDec
	}{
		{
			"no epochs per period",
			sdk.TokensFromConsensusPower(400_000_000, evmostypes.PowerReduction).Sub(bondedAmt),
			func() {
				suite.app.InflationKeeper.SetEpochsPerPeriod(suite.ctx, 0)
			},
			math.LegacyZeroDec(),
		},
		{
			"high supply",
			sdk.TokensFromConsensusPower(800_000_000, evmostypes.PowerReduction).Sub(bondedAmt),
			func() {},
			math.LegacyMustNewDecFromStr("17.187500000000000000"),
		},
		{
			"low supply",
			sdk.TokensFromConsensusPower(400_000_000, evmostypes.PowerReduction).Sub(bondedAmt),
			func() {},
			math.LegacyMustNewDecFromStr("51.562500000000000000"),
		},
		{
			"zero circulating supply",
			sdk.TokensFromConsensusPower(200_000_000, evmostypes.PowerReduction).Sub(bondedAmt),
			func() {},
			math.LegacyZeroDec(),
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			// Team allocation is only set on mainnet
			suite.ctx = suite.ctx.WithChainID("evmos_9001-1")
			tc.malleate()

			// Mint coins to increase supply
			coin := sdk.NewCoin(
				types.DefaultInflationDenom,
				tc.bankSupply,
			)
			decCoin := sdk.NewDecCoinFromCoin(coin)
			err := suite.app.InflationKeeper.MintCoins(suite.ctx, coin)
			suite.Require().NoError(err)

			teamAlloc := sdk.NewDecCoin(
				types.DefaultInflationDenom,
				sdk.TokensFromConsensusPower(int64(200_000_000), evmostypes.PowerReduction),
			)

			circulatingSupply := s.app.InflationKeeper.GetCirculatingSupply(suite.ctx, types.DefaultInflationDenom)
			suite.Require().Equal(decCoin.Add(bondedCoins).Sub(teamAlloc).Amount, circulatingSupply)

			inflationRate := s.app.InflationKeeper.GetInflationRate(suite.ctx, types.DefaultInflationDenom)
			suite.Require().Equal(tc.expInflationRate, inflationRate)
		})
	}
}

func (suite *KeeperTestSuite) TestBondedRatio() {
	testCases := []struct {
		name         string
		isMainnet    bool
		malleate     func()
		expBondRatio math.LegacyDec
	}{
		{
			"is mainnet",
			true,
			func() {},
			math.LegacyZeroDec(),
		},
		{
			"not mainnet",
			false,
			func() {},
			math.LegacyMustNewDecFromStr("0.999900009999000099"),
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			// Team allocation is only set on mainnet
			if tc.isMainnet {
				suite.ctx = suite.ctx.WithChainID("evmos_9001-1")
			} else {
				suite.ctx = suite.ctx.WithChainID("evmos_9999-666")
			}
			tc.malleate()

			bondRatio := suite.app.InflationKeeper.BondedRatio(suite.ctx)
			suite.Require().Equal(tc.expBondRatio, bondRatio)
		})
	}
}
