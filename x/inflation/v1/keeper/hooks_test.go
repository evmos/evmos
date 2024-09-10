package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	epochstypes "github.com/evmos/evmos/v19/x/epochs/types"
	"github.com/evmos/evmos/v19/x/inflation/v1/types"
	"github.com/stretchr/testify/require"
)

func TestEpochIdentifierAfterEpochEnd(t *testing.T) {
	var (
		ctx sdk.Context
		nw  *network.UnitTestNetwork
	)
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
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			// reset
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()

			params := nw.App.InflationKeeper.GetParams(ctx)
			params.EnableInflation = true
			err := nw.App.InflationKeeper.SetParams(ctx, params)
			require.NoError(t, err)

			futureCtx := ctx.WithBlockTime(time.Now().Add(time.Hour))
			newHeight := nw.App.LastBlockHeight() + 1

			feePoolOrigin, err := nw.App.DistrKeeper.FeePool.Get(ctx)
			require.NoError(t, err)
			nw.App.EpochsKeeper.BeforeEpochStart(futureCtx, tc.epochIdentifier, newHeight)
			nw.App.EpochsKeeper.AfterEpochEnd(futureCtx, tc.epochIdentifier, newHeight)

			nw.App.EpochsKeeper.AfterEpochEnd(futureCtx, tc.epochIdentifier, newHeight)

			// check the distribution happened as well
			feePoolNew, err := nw.App.DistrKeeper.FeePool.Get(ctx)
			require.NoError(t, err)
			if tc.expDistribution {
				// Actual distribution portions are tested elsewhere; we just want to verify the value of the pool is greater here
				require.Greater(t, feePoolNew.CommunityPool.AmountOf(denomMint).BigInt().Uint64(),
					feePoolOrigin.CommunityPool.AmountOf(denomMint).BigInt().Uint64())
			} else {
				require.Equal(t, feePoolNew.CommunityPool.AmountOf(denomMint), feePoolOrigin.CommunityPool.AmountOf(denomMint))
			}
		})
	}
}

func TestPeriodChangesSkippedEpochsAfterEpochEnd(t *testing.T) {
	nw := network.NewUnitTestNetwork()
	ctx := nw.GetContext()

	currentEpochPerPeriod := nw.App.InflationKeeper.GetEpochsPerPeriod(ctx)
	// bondingRatio is zero in tests
	bondedRatio, err := nw.App.InflationKeeper.BondedRatio(ctx)
	require.NoError(t, err)
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
			currentEpochPerPeriod - 10, // so it's within range
			epochstypes.DayEpochID,
			0,
			false,
			false,
		},
		{
			"SkippedEpoch set WeekEpochID disabledInflation ",
			0,
			currentEpochPerPeriod - 10, // so it's within range
			epochstypes.WeekEpochID,
			0,
			false,
			false,
		},
		{
			"[Period 0] disabledInflation",
			0,
			currentEpochPerPeriod - 10, // so it's within range
			epochstypes.DayEpochID,
			0,
			false,
			false,
		},
		{
			"[Period 0] period stays the same under epochs per period",
			0,
			currentEpochPerPeriod - 10, // so it's within range
			epochstypes.DayEpochID,
			0,
			true,
			false,
		},
		{
			"[Period 0] period changes once enough epochs have passed",
			0,
			currentEpochPerPeriod + 1,
			epochstypes.DayEpochID,
			0,
			true,
			true,
		},
		{
			"[Period 1] period stays the same under the epoch per period",
			1,
			2*currentEpochPerPeriod - 1,
			epochstypes.DayEpochID,
			0,
			true,
			false,
		},
		{
			"[Period 1] period changes once enough epochs have passed",
			1,
			2*currentEpochPerPeriod + 1,
			epochstypes.DayEpochID,
			0,
			true,
			true,
		},
		{
			"[Period 0] with skipped epochs - period stays the same under epochs per period",
			0,
			currentEpochPerPeriod - 1,
			epochstypes.DayEpochID,
			10,
			true,
			false,
		},
		{
			"[Period 0] with skipped epochs - period stays the same under epochs per period",
			0,
			currentEpochPerPeriod + 1,
			epochstypes.DayEpochID,
			10,
			true,
			false,
		},
		{
			"[Period 0] with skipped epochs - period changes once enough epochs have passed",
			0,
			currentEpochPerPeriod + 11,
			epochstypes.DayEpochID,
			10,
			true,
			true,
		},
		{
			"[Period 1] with skipped epochs - period stays the same under epochs per period",
			1,
			2*currentEpochPerPeriod + 1,
			epochstypes.DayEpochID,
			10,
			true,
			false,
		},
		{
			"[Period 1] with skipped epochs - period changes once enough epochs have passed",
			1,
			2*currentEpochPerPeriod + 11,
			epochstypes.DayEpochID,
			10,
			true,
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			// reset
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()

			params := nw.App.InflationKeeper.GetParams(ctx)
			params.EnableInflation = true
			err := nw.App.InflationKeeper.SetParams(ctx, params)
			require.NoError(t, err)

			// Before hook
			if !tc.enableInflation {
				params.EnableInflation = false
				err = nw.App.InflationKeeper.SetParams(ctx, params)
				require.NoError(t, err)
			}

			nw.App.InflationKeeper.SetSkippedEpochs(ctx, tc.skippedEpochs)
			nw.App.InflationKeeper.SetPeriod(ctx, uint64(tc.currentPeriod)) //nolint:gosec // G115
			currentSkippedEpochs := nw.App.InflationKeeper.GetSkippedEpochs(ctx)
			currentPeriod := nw.App.InflationKeeper.GetPeriod(ctx)
			originalProvision := nw.App.InflationKeeper.GetEpochMintProvision(ctx)

			// Perform Epoch Hooks
			futureCtx := ctx.WithBlockTime(time.Now().Add(time.Minute))
			nw.App.EpochsKeeper.BeforeEpochStart(futureCtx, tc.epochIdentifier, tc.height)
			nw.App.EpochsKeeper.AfterEpochEnd(futureCtx, tc.epochIdentifier, tc.height)
			skippedEpochs := nw.App.InflationKeeper.GetSkippedEpochs(ctx)
			period := nw.App.InflationKeeper.GetPeriod(ctx)

			if tc.periodChanges {
				newProvision := nw.App.InflationKeeper.GetEpochMintProvision(ctx)
				expectedProvision := types.CalculateEpochMintProvision(
					nw.App.InflationKeeper.GetParams(ctx),
					period,
					currentEpochPerPeriod,
					bondedRatio,
				).Quo(math.LegacyNewDec(types.ReductionFactor))
				require.Equal(t, expectedProvision, newProvision)
				// mint provisions will change
				require.NotEqual(t, newProvision.BigInt().Uint64(), originalProvision.BigInt().Uint64())
				require.Equal(t, currentSkippedEpochs, skippedEpochs)
				require.Equal(t, currentPeriod+1, period)
			} else {
				require.Equal(t, currentPeriod, period)
				if !tc.enableInflation {
					// Check for epochIdentifier for skippedEpoch increment
					if tc.epochIdentifier == epochstypes.DayEpochID {
						require.Equal(t, currentSkippedEpochs+1, skippedEpochs)
					}
				}
			}
		})
	}
}
