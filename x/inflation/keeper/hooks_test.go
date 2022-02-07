package keeper_test

import (
	"fmt"
	"time"

	"github.com/tharsis/evmos/x/inflation/types"
)

func (suite *KeeperTestSuite) TestEpochIdentifierAfterEpochEnd() {
	testCases := []struct {
		name            string
		epochIdentifier string
		expDistribution bool
	}{
		{
			"correct epoch identifier",
			"day",
			true,
		},
		{
			"incorrect epoch identifier",
			"week",
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()

			futureCtx := suite.ctx.WithBlockTime(time.Now().Add(time.Hour))
			newHeight := suite.app.LastBlockHeight() + 1

			feePoolOrigin := suite.app.DistrKeeper.GetFeePool(suite.ctx)
			suite.app.EpochsKeeper.BeforeEpochStart(futureCtx, tc.epochIdentifier, newHeight)
			suite.app.EpochsKeeper.AfterEpochEnd(futureCtx, tc.epochIdentifier, newHeight)

			suite.app.EpochsKeeper.AfterEpochEnd(futureCtx, tc.epochIdentifier, newHeight)

			// check the distribution happened as well
			feePoolNew := suite.app.DistrKeeper.GetFeePool(suite.ctx)
			if tc.expDistribution {
				// Actual distribution portions are tested elsewhere; we just want to verify the value of the pool is greater here
				suite.Require().Greater(feePoolNew.CommunityPool.AmountOf(denomMint).BigInt().Uint64(),
					feePoolOrigin.CommunityPool.AmountOf(denomMint).BigInt().Uint64())
			} else {
				suite.Require().Equal(feePoolNew.CommunityPool.AmountOf(denomMint), feePoolOrigin.CommunityPool.AmountOf(denomMint))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestPeriodChangesAfterEpochEnd() {
	suite.SetupTest()

	currentEpochPeriod := suite.app.InflationKeeper.GetEpochsPerPeriod(suite.ctx)
	// bondingRatio is zero in tests
	bondedRatio := suite.app.StakingKeeper.BondedRatio(suite.ctx)

	testCases := []struct {
		name    string
		height  int64
		changes bool
	}{
		{
			"[Period 0] period stays the same under epoch per period",
			currentEpochPeriod - 10, // so it's within range
			false,
		},
		{
			"[Period 0] period changes once enough epochs have passed",
			currentEpochPeriod + 1,
			true,
		},
		{
			"[Period 1] period stays the same under the epoch per period",
			2*currentEpochPeriod - 1,
			false,
		},
		{
			"[Period 1] period changes once enough epochs have passed",
			2*currentEpochPeriod + 1,
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			currentPeriod := suite.app.InflationKeeper.GetPeriod(suite.ctx)
			epochIdentifier := suite.app.InflationKeeper.GetEpochIdentifier(suite.ctx)

			originalProvision, found := suite.app.InflationKeeper.GetEpochMintProvision(suite.ctx)
			suite.Require().True(found)

			futureCtx := suite.ctx.WithBlockTime(time.Now().Add(time.Minute))
			suite.app.EpochsKeeper.BeforeEpochStart(futureCtx, epochIdentifier, tc.height)
			suite.app.EpochsKeeper.AfterEpochEnd(futureCtx, epochIdentifier, tc.height)
			newPeriod := suite.app.InflationKeeper.GetPeriod(suite.ctx)

			if tc.changes {
				newProvision, found := suite.app.InflationKeeper.GetEpochMintProvision(suite.ctx)
				suite.Require().True(found)
				expectedProvision := types.CalculateEpochMintProvision(
					suite.app.InflationKeeper.GetParams(suite.ctx),
					newPeriod,
					currentEpochPeriod,
					bondedRatio,
				)
				suite.Require().Equal(expectedProvision, newProvision)
				// mint provisions will change
				suite.Require().NotEqual(newProvision.BigInt().Uint64(), originalProvision.BigInt().Uint64())
				suite.Require().Equal(currentPeriod+1, newPeriod)
			} else {
				suite.Require().Equal(newPeriod, currentPeriod)
			}
		})
	}
}
