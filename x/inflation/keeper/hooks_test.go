package keeper_test

import (
	"fmt"
	"time"

	epochstypes "github.com/evmos/evmos/v6/x/epochs/types"
	"github.com/evmos/evmos/v6/x/inflation/types"
)

func (suite *KeeperTestSuite) TestEpochIdentifierAfterEpochEnd() {
	testCases := []struct {
		name            string
		epochIdentifier string
		expDistribution bool
	}{
		{
			"correct epoch identifier",
			epochstypes.DayEpochID,
			true,
		},
		{
			"incorrect epoch identifier",
			epochstypes.WeekEpochID,
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()

			params := suite.app.InflationKeeper.GetParams(suite.ctx)
			params.EnableInflation = true
			suite.app.InflationKeeper.SetParams(suite.ctx, params)

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

func (suite *KeeperTestSuite) TestPeriodChangesSkippedEpochsAfterEpochEnd() {
	suite.SetupTest()

	currentEpochPeriod := suite.app.InflationKeeper.GetEpochsPerPeriod(suite.ctx)
	// bondingRatio is zero in tests
	bondedRatio := suite.app.InflationKeeper.BondedRatio(suite.ctx)

	testCases := []struct {
		name            string
		currentPeriod   int64
		height          int64
		epochIdentifier string
		skippedEpochs   uint64
		enableInflation bool
		periodChanges   bool
	}{
		{
			"SkippedEpoch set DayEpochID disabledInflation",
			0,
			currentEpochPeriod - 10, // so it's within range
			epochstypes.DayEpochID,
			0,
			false,
			false,
		},
		{
			"SkippedEpoch set WeekEpochID disabledInflation ",
			0,
			currentEpochPeriod - 10, // so it's within range
			epochstypes.WeekEpochID,
			0,
			false,
			false,
		},
		{
			"[Period 0] disabledInflation",
			0,
			currentEpochPeriod - 10, // so it's within range
			epochstypes.DayEpochID,
			0,
			false,
			false,
		},
		{
			"[Period 0] period stays the same under epochs per period",
			0,
			currentEpochPeriod - 10, // so it's within range
			epochstypes.DayEpochID,
			0,
			true,
			false,
		},
		{
			"[Period 0] period changes once enough epochs have passed",
			0,
			currentEpochPeriod + 1,
			epochstypes.DayEpochID,
			0,
			true,
			true,
		},
		{
			"[Period 1] period stays the same under the epoch per period",
			1,
			2*currentEpochPeriod - 1,
			epochstypes.DayEpochID,
			0,
			true,
			false,
		},
		{
			"[Period 1] period changes once enough epochs have passed",
			1,
			2*currentEpochPeriod + 1,
			epochstypes.DayEpochID,
			0,
			true,
			true,
		},
		{
			"[Period 0] with skipped epochs - period stays the same under epochs per period",
			0,
			currentEpochPeriod - 1,
			epochstypes.DayEpochID,
			10,
			true,
			false,
		},
		{
			"[Period 0] with skipped epochs - period stays the same under epochs per period",
			0,
			currentEpochPeriod + 1,
			epochstypes.DayEpochID,
			10,
			true,
			false,
		},
		{
			"[Period 0] with skipped epochs - period changes once enough epochs have passed",
			0,
			currentEpochPeriod + 11,
			epochstypes.DayEpochID,
			10,
			true,
			true,
		},
		{
			"[Period 1] with skipped epochs - period stays the same under epochs per period",
			1,
			2*currentEpochPeriod + 1,
			epochstypes.DayEpochID,
			10,
			true,
			false,
		},
		{
			"[Period 1] with skipped epochs - period changes once enough epochs have passed",
			1,
			2*currentEpochPeriod + 11,
			epochstypes.DayEpochID,
			10,
			true,
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			params := suite.app.InflationKeeper.GetParams(suite.ctx)
			params.EnableInflation = true
			suite.app.InflationKeeper.SetParams(suite.ctx, params)

			// Before hook
			if !tc.enableInflation {
				params.EnableInflation = false
				suite.app.InflationKeeper.SetParams(suite.ctx, params)
			}

			suite.app.InflationKeeper.SetSkippedEpochs(suite.ctx, tc.skippedEpochs)
			suite.app.InflationKeeper.SetPeriod(suite.ctx, uint64(tc.currentPeriod))
			currentSkippedEpochs := suite.app.InflationKeeper.GetSkippedEpochs(suite.ctx)
			currentPeriod := suite.app.InflationKeeper.GetPeriod(suite.ctx)
			originalProvision, found := suite.app.InflationKeeper.GetEpochMintProvision(suite.ctx)
			suite.Require().True(found)

			// Perform Epoch Hooks
			futureCtx := suite.ctx.WithBlockTime(time.Now().Add(time.Minute))
			suite.app.EpochsKeeper.BeforeEpochStart(futureCtx, tc.epochIdentifier, tc.height)
			suite.app.EpochsKeeper.AfterEpochEnd(futureCtx, tc.epochIdentifier, tc.height)
			skippedEpochs := suite.app.InflationKeeper.GetSkippedEpochs(suite.ctx)
			period := suite.app.InflationKeeper.GetPeriod(suite.ctx)

			if tc.periodChanges {
				newProvision, found := suite.app.InflationKeeper.GetEpochMintProvision(suite.ctx)
				suite.Require().True(found)
				expectedProvision := types.CalculateEpochMintProvision(
					suite.app.InflationKeeper.GetParams(suite.ctx),
					period,
					currentEpochPeriod,
					bondedRatio,
				)
				suite.Require().Equal(expectedProvision, newProvision)
				// mint provisions will change
				suite.Require().NotEqual(newProvision.BigInt().Uint64(), originalProvision.BigInt().Uint64())
				suite.Require().Equal(currentSkippedEpochs, skippedEpochs)
				suite.Require().Equal(currentPeriod+1, period)
			} else {
				suite.Require().Equal(currentPeriod, period)
				if !tc.enableInflation {
					// Check for epochIdentifier for skippedEpoch increment
					if tc.epochIdentifier == epochstypes.DayEpochID {
						suite.Require().Equal(currentSkippedEpochs+1, skippedEpochs)
					}
				}
			}
		})
	}
}
