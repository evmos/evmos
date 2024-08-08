package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/evmos/evmos/v19/x/epochs/types"
)

func TestEpochInfoChangesBeginBlockerAndInitGenesis(t *testing.T) {
	var (
		suite     *KeeperTestSuite
		epochInfo types.EpochInfo
		found     bool
		now       = time.Now().UTC()
	)

	testCases := []struct {
		name                       string
		expCountingStarted         bool
		expCurrentEpochStartTime   time.Time
		expCurrentEpochStartHeight int64
		expCurrentEpoch            int64
		expInitialEpochStartTime   time.Time
		malleate                   func()
	}{
		{
			name:                       "pass - initial epoch not started",
			expCountingStarted:         false,
			expCurrentEpochStartHeight: 0,
			expCurrentEpochStartTime:   now,
			expCurrentEpoch:            0,
			expInitialEpochStartTime:   now.Add(time.Second),
			malleate: func() {
				ctx := suite.network.GetContext()
				epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(ctx, monthIdentifier)
				require.True(t, found)
			},
		},
		{
			// We are assuming a block time of 1 second here. The first block is created during
			// suite initialization so here we are at the second block.
			name:                       "pass - initial epoch started",
			expCountingStarted:         true,
			expCurrentEpochStartHeight: 2,
			expCurrentEpochStartTime:   now.Add(time.Second),
			expCurrentEpoch:            1,
			expInitialEpochStartTime:   now.Add(time.Second),
			malleate: func() {
				ctx := suite.network.GetContext()
				ctx = ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				require.NoError(t, suite.network.App.EpochsKeeper.BeginBlocker(ctx))
				epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(ctx, monthIdentifier)
				require.True(t, found)
			},
		},
		{
			name:                       "pass - second epoch started",
			expCountingStarted:         true,
			expCurrentEpochStartHeight: 3,
			expCurrentEpochStartTime:   now.Add(time.Second).Add(month),
			expCurrentEpoch:            2,
			expInitialEpochStartTime:   now.Add(time.Second),
			malleate: func() {
				ctx := suite.network.GetContext()
				// Epoch start
				ctx = ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				require.NoError(t, suite.network.App.EpochsKeeper.BeginBlocker(ctx))
				// Here we use seconds * 2 because we have to be 1 second more the end of previous
				// epoch.
				ctx = ctx.WithBlockHeight(3).WithBlockTime(now.Add(2 * time.Second).Add(month))
				require.NoError(t, suite.network.App.EpochsKeeper.BeginBlocker(ctx))
				epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(ctx, monthIdentifier)
				require.True(t, found)
			},
		},
		{
			name:                       "pass - still second epoch adding 1 month to epoch start",
			expCountingStarted:         true,
			expCurrentEpochStartHeight: 3,
			expCurrentEpochStartTime:   now.Add(time.Second).Add(month),
			expCurrentEpoch:            2,
			expInitialEpochStartTime:   now.Add(time.Second),
			malleate: func() {
				ctx := suite.network.GetContext()
				ctx = ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				require.NoError(t, suite.network.App.EpochsKeeper.BeginBlocker(ctx))
				ctx = ctx.WithBlockHeight(3).WithBlockTime(now.Add(2 * time.Second).Add(month))
				require.NoError(t, suite.network.App.EpochsKeeper.BeginBlocker(ctx))
				ctx = ctx.WithBlockHeight(4).WithBlockTime(now.Add(time.Second).Add(2 * month))
				require.NoError(t, suite.network.App.EpochsKeeper.BeginBlocker(ctx))
				epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(ctx, monthIdentifier)
				require.True(t, found)
			},
		},
		{
			name:                       "pass - third epoch start 1 month plus 1 second from previous epoch start",
			expCountingStarted:         true,
			expCurrentEpochStartHeight: 4,
			// NOTE: Even though previous epoch to complete needs 1 second more than its end,
			// the start of next one is stored as equal to previous epoch end.
			expCurrentEpochStartTime: now.Add(time.Second).Add(2 * month),
			expCurrentEpoch:          3,
			expInitialEpochStartTime: now.Add(time.Second),
			malleate: func() {
				ctx := suite.network.GetContext()
				ctx = ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				require.NoError(t, suite.network.App.EpochsKeeper.BeginBlocker(ctx))
				ctx = ctx.WithBlockHeight(3).WithBlockTime(now.Add(2 * time.Second).Add(month))
				require.NoError(t, suite.network.App.EpochsKeeper.BeginBlocker(ctx))
				ctx = ctx.WithBlockHeight(4).WithBlockTime(now.Add(2 * time.Second).Add(2 * month))
				require.NoError(t, suite.network.App.EpochsKeeper.BeginBlocker(ctx))
				epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(ctx, monthIdentifier)
				require.True(t, found)
			},
		},
		{
			name:                       "pass - still third epoch adding 1 day from start",
			expCountingStarted:         true,
			expCurrentEpochStartHeight: 4,
			expCurrentEpochStartTime:   now.Add(time.Second).Add(2 * month),
			expCurrentEpoch:            3,
			expInitialEpochStartTime:   now.Add(time.Second),
			malleate: func() {
				ctx := suite.network.GetContext()
				// First epoch
				ctx = ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				require.NoError(t, suite.network.App.EpochsKeeper.BeginBlocker(ctx))
				// Second epoch
				ctx = ctx.WithBlockHeight(3).WithBlockTime(now.Add(2 * time.Second).Add(month))
				require.NoError(t, suite.network.App.EpochsKeeper.BeginBlocker(ctx))
				// Third epoch
				ctx = ctx.WithBlockHeight(4).WithBlockTime(now.Add(2 * time.Second).Add(2 * month))
				require.NoError(t, suite.network.App.EpochsKeeper.BeginBlocker(ctx))
				// Still third epoch
				ctx = ctx.WithBlockHeight(5).WithBlockTime(now.Add(2 * time.Second).Add(2 * month).Add(day))
				require.NoError(t, suite.network.App.EpochsKeeper.BeginBlocker(ctx))
				epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(ctx, monthIdentifier)
				require.True(t, found)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			// custom genesis defines an epoch that is not yet start but that should start at
			// 1 second after the genesis time equal to now. This happens in the BeginBlocker.
			epochsInfo := []types.EpochInfo{
				{
					Identifier:              monthIdentifier,
					StartTime:               now.Add(time.Second),
					Duration:                month,
					CurrentEpoch:            0,
					CurrentEpochStartHeight: 0,
					CurrentEpochStartTime:   now,
					EpochCountingStarted:    false,
				},
			}
			suite = SetupTest(epochsInfo)

			// Check that custom genesis worked.
			epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(suite.network.GetContext(), monthIdentifier)
			require.True(t, found, "expected to find custom genesis data")

			tc.malleate()

			require.Equal(t, monthIdentifier, epochInfo.Identifier, "expected a different identifier")
			require.Equal(t, month, epochInfo.Duration, "expected a different duration")
			require.Equal(t, tc.expCurrentEpoch, epochInfo.CurrentEpoch, "expected a different current epoch")
			require.Equal(t, tc.expCurrentEpochStartHeight, epochInfo.CurrentEpochStartHeight, "expected different current epoch start height")
			require.Equal(t, tc.expCurrentEpochStartTime.UTC().String(), epochInfo.CurrentEpochStartTime.UTC().String(), "expected different current epoch start time")
			require.Equal(t, tc.expCountingStarted, epochInfo.EpochCountingStarted, "expected different epoch counting started")
			require.Equal(t, tc.expInitialEpochStartTime.UTC().String(), epochInfo.StartTime.UTC().String(), "expected a different initial start time")
		})
	}
}

func TestEpochStartingOneMonthAfterInitGenesis(t *testing.T) {
	now := time.Now().UTC()
	nowPlusMonth := now.Add(month)

	epochsInfo := []types.EpochInfo{
		{
			Identifier:              monthIdentifier,
			StartTime:               now.Add(month),
			Duration:                month,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   now,
			EpochCountingStarted:    false,
		},
	}
	suite := SetupTest(epochsInfo)
	ctx := suite.network.GetContext()

	// Epoch not started yet.
	epochInfo, found := suite.network.App.EpochsKeeper.GetEpochInfo(ctx, monthIdentifier)
	require.True(t, found)
	require.Equal(t, int64(0), epochInfo.CurrentEpoch, "expected first epoch not started")
	require.Equal(t, int64(0), epochInfo.CurrentEpochStartHeight, "expected current epoch start height 0")
	require.Equal(t, now, epochInfo.CurrentEpochStartTime, "expected current epoch start time equal to genesis time.")
	require.Equal(t, false, epochInfo.EpochCountingStarted, "expected epoch counting not started")

	// After 1 week.
	ctx = ctx.WithBlockHeight(2).WithBlockTime(now.Add(week))
	require.NoError(t, suite.network.App.EpochsKeeper.BeginBlocker(ctx))

	epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(ctx, monthIdentifier)
	require.True(t, found)
	require.Equal(t, int64(0), epochInfo.CurrentEpoch, "expected first epoch not started")
	require.Equal(t, int64(0), epochInfo.CurrentEpochStartHeight, "expected current epoch start height 0")
	require.Equal(t, now, epochInfo.CurrentEpochStartTime, "expected current epoch start time equal to genesis time.")
	require.Equal(t, false, epochInfo.EpochCountingStarted, "expected epoch counting not started")

	// After 1 month.
	ctx = ctx.WithBlockHeight(3).WithBlockTime(nowPlusMonth)
	require.NoError(t, suite.network.App.EpochsKeeper.BeginBlocker(ctx))

	// epoch started
	epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(ctx, monthIdentifier)
	require.True(t, found)
	require.Equal(t, int64(1), epochInfo.CurrentEpoch, "expected current epoch equal to first epoch")
	require.Equal(t, ctx.BlockHeight(), epochInfo.CurrentEpochStartHeight, "expected current epoch start height equal to current height")
	require.Equal(t, nowPlusMonth.UTC().String(), epochInfo.CurrentEpochStartTime.UTC().String(), "expected a different start time for the epoch")
	require.Equal(t, true, epochInfo.EpochCountingStarted, "expected epoch counting started")
}
