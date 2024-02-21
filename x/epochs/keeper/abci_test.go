package keeper_test

import (
	"fmt"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v16/x/epochs/types"
)

func (suite *KeeperTestSuite) TestEpochInfoChangesBeginBlockerAndInitGenesis() {
	var (
		epochInfo types.EpochInfo
		found     bool
		ctx       sdktypes.Context
	)
	day := time.Hour * 24
	month := time.Hour * 24 * 31

	testCases := []struct {
		name                       string
		expCountingStarted         bool
		expCurrentEpochStartTime   time.Time
		expCurrentEpochStartHeight int64
		expCurrentEpoch            int64
		expInitialEpochStartTime   time.Time
		malleate                   func(ctx sdktypes.Context)
	}{
		{
			name:                       "pass - initial epoch not started",
			expCountingStarted:         false,
			expCurrentEpochStartHeight: 0,
			expCurrentEpochStartTime:   time.Time{},
			expCurrentEpoch:            0,
			expInitialEpochStartTime:   time.Time{}.Add(time.Second),
			malleate: func(ctx sdktypes.Context) {
				epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(ctx, "month")
				suite.Require().True(found)
			},
		},
		{
			// We are assuming a block time of 1 second here. The first block is created during
			// suite initialization so here we are at the second block.
			name:                       "pass - initial epoch started",
			expCountingStarted:         true,
			expCurrentEpochStartHeight: 2,
			expCurrentEpochStartTime:   time.Time{}.Add(time.Second),
			expCurrentEpoch:            1,
			expInitialEpochStartTime:   time.Time{}.Add(time.Second),
			malleate: func(ctx sdktypes.Context) {
				ctx = ctx.WithBlockHeight(2).WithBlockTime(time.Time{}.Add(time.Second))
				suite.network.App.EpochsKeeper.BeginBlocker(ctx)
				epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(ctx, "month")
				suite.Require().True(found)
			},
		},
		{
			name:                       "pass - second epoch started",
			expCountingStarted:         true,
			expCurrentEpochStartHeight: 3,
			expCurrentEpochStartTime:   time.Time{}.Add(time.Second).Add(month),
			expCurrentEpoch:            2,
			expInitialEpochStartTime:   time.Time{}.Add(time.Second),
			malleate: func(ctx sdktypes.Context) {
				// Epoch start
				ctx = ctx.WithBlockHeight(2).WithBlockTime(time.Time{}.Add(time.Second))
				suite.network.App.EpochsKeeper.BeginBlocker(ctx)
				// Here we use seconds * 2 because we have to be 1 second more the end of previous
				// epoch.
				ctx = ctx.WithBlockHeight(3).WithBlockTime(time.Time{}.Add(2 * time.Second).Add(month))
				suite.network.App.EpochsKeeper.BeginBlocker(ctx)
				epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(ctx, "month")
				fmt.Printf("%+v\n", epochInfo)
				suite.Require().True(found)
			},
		},
		{
			name:                       "pass - still second epoch adding 1 month to epoch start",
			expCountingStarted:         true,
			expCurrentEpochStartHeight: 3,
			expCurrentEpochStartTime:   time.Time{}.Add(time.Second).Add(month),
			expCurrentEpoch:            2,
			expInitialEpochStartTime:   time.Time{}.Add(time.Second),
			malleate: func(ctx sdktypes.Context) {
				ctx = ctx.WithBlockHeight(2).WithBlockTime(time.Time{}.Add(time.Second))
				suite.network.App.EpochsKeeper.BeginBlocker(ctx)
				ctx = ctx.WithBlockHeight(3).WithBlockTime(time.Time{}.Add(2 * time.Second).Add(month))
				suite.network.App.EpochsKeeper.BeginBlocker(ctx)
				ctx = ctx.WithBlockHeight(4).WithBlockTime(time.Time{}.Add(time.Second).Add(2 * month))
				suite.network.App.EpochsKeeper.BeginBlocker(ctx)
				epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(ctx, "month")
				suite.Require().True(found)
			},
		},
		{
			name:                       "pass - third epoch start 1 month plus 1 second from previous epoch start",
			expCountingStarted:         true,
			expCurrentEpochStartHeight: 4,
			// Even though previous epoch to complete we need 1 second more than its end,
			// the start of new epoch is stored as equal to previous epoch end.
			expCurrentEpochStartTime: time.Time{}.Add(time.Second).Add(2 * month),
			expCurrentEpoch:          3,
			expInitialEpochStartTime: time.Time{}.Add(time.Second),
			malleate: func(ctx sdktypes.Context) {
				ctx = ctx.WithBlockHeight(2).WithBlockTime(time.Time{}.Add(time.Second))
				suite.network.App.EpochsKeeper.BeginBlocker(ctx)
				ctx = ctx.WithBlockHeight(3).WithBlockTime(time.Time{}.Add(2 * time.Second).Add(month))
				suite.network.App.EpochsKeeper.BeginBlocker(ctx)
				ctx = ctx.WithBlockHeight(4).WithBlockTime(time.Time{}.Add(2 * time.Second).Add(2 * month))
				suite.network.App.EpochsKeeper.BeginBlocker(ctx)
				epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(ctx, "month")
				suite.Require().True(found)
			},
		},
		{ //nolint:dupl
			name:                       "pass - still third epoch adding 1 day from start",
			expCountingStarted:         true,
			expCurrentEpochStartHeight: 4,
			expCurrentEpochStartTime: time.Time{}.Add(time.Second).Add(2 * month),
			expCurrentEpoch:            3,
			expInitialEpochStartTime: time.Time{}.Add(time.Second),
			malleate: func(ctx sdktypes.Context) {
                // First epoch
				ctx = ctx.WithBlockHeight(2).WithBlockTime(time.Time{}.Add(time.Second))
				suite.network.App.EpochsKeeper.BeginBlocker(ctx)
                // Second epoch
				ctx = ctx.WithBlockHeight(3).WithBlockTime(time.Time{}.Add(2 * time.Second).Add(month))
				suite.network.App.EpochsKeeper.BeginBlocker(ctx)
                // Third epoch
				ctx = ctx.WithBlockHeight(4).WithBlockTime(time.Time{}.Add(2 * time.Second).Add(2 * month))
				suite.network.App.EpochsKeeper.BeginBlocker(ctx)
                // Still third epoch
				ctx = ctx.WithBlockHeight(5).WithBlockTime(time.Time{}.Add(2 * time.Second).Add(2 * month).Add(day))
				suite.network.App.EpochsKeeper.BeginBlocker(ctx)
				epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(ctx, "month")
				suite.Require().True(found)
			},
		},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %d", i), func() {
			identifier := "month"

            // custom genesis defines an epoch that is not yet start but that should start at
            // specific time and block. This should happen in the BeginBlocker.
			epochsInfo := []types.EpochInfo{
				{
					Identifier:              identifier,
					StartTime:               time.Time{}.Add(time.Second),
					Duration:                month,
					CurrentEpoch:            0,
					CurrentEpochStartHeight: 0,
					CurrentEpochStartTime:   time.Time{},
					EpochCountingStarted:    false,
				},
			}
			ctx = suite.SetupTest(epochsInfo) // reset

			// Check that custom genesis worked.
			epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(ctx, "month")
			suite.Require().True(found, "expected to find custom genesis data")

			tc.malleate(ctx)

			suite.Require().Equal(identifier, epochInfo.Identifier, "expected a different identifier")
			suite.Require().Equal(month, epochInfo.Duration, "expected a different duration")
			suite.Require().Equal(tc.expCurrentEpoch, epochInfo.CurrentEpoch, "expected a different current epoch")
			suite.Require().Equal(tc.expCurrentEpochStartHeight, epochInfo.CurrentEpochStartHeight, "expected different current epoch start height")
			suite.Require().Equal(tc.expCurrentEpochStartTime.UTC().String(), epochInfo.CurrentEpochStartTime.UTC().String(), "expected different current epoch start time")
			suite.Require().Equal(tc.expCountingStarted, epochInfo.EpochCountingStarted, "expected different epoch counting started")
			suite.Require().Equal(tc.expInitialEpochStartTime.UTC().String(), epochInfo.StartTime.UTC().String(), "expected a different initial start time")
		})
	}
}

// func (suite *KeeperTestSuite) TestEpochStartingOneMonthAfterInitGenesis() {
// 	// On init genesis, default epochs information is set
// 	// To check init genesis again, should make it fresh status
// 	epochInfos := suite.app.EpochsKeeper.AllEpochInfos(suite.ctx)
// 	for _, epochInfo := range epochInfos {
// 		suite.app.EpochsKeeper.DeleteEpochInfo(suite.ctx, epochInfo.Identifier)
// 	}
//
// 	now := time.Now()
// 	week := time.Hour * 24 * 7
// 	month := time.Hour * 24 * 30
// 	initialBlockHeight := int64(1)
// 	suite.ctx = suite.ctx.WithBlockHeight(initialBlockHeight).WithBlockTime(now)
//
// 	epochs.InitGenesis(suite.ctx, suite.app.EpochsKeeper, types.GenesisState{
// 		Epochs: []types.EpochInfo{
// 			{
// 				Identifier:              "monthly",
// 				StartTime:               now.Add(month),
// 				Duration:                time.Hour * 24 * 30,
// 				CurrentEpoch:            0,
// 				CurrentEpochStartHeight: suite.ctx.BlockHeight(),
// 				CurrentEpochStartTime:   time.Time{},
// 				EpochCountingStarted:    false,
// 			},
// 		},
// 	})
//
// 	// epoch not started yet
// 	epochInfo, found := suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "monthly")
// 	suite.Require().True(found)
// 	suite.Require().Equal(epochInfo.CurrentEpoch, int64(0))
// 	suite.Require().Equal(epochInfo.CurrentEpochStartHeight, initialBlockHeight)
// 	suite.Require().Equal(epochInfo.CurrentEpochStartTime, time.Time{})
// 	suite.Require().Equal(epochInfo.EpochCountingStarted, false)
//
// 	// after 1 week
// 	suite.ctx = suite.ctx.WithBlockHeight(2).WithBlockTime(now.Add(week))
// 	suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
//
// 	// epoch not started yet
// 	epochInfo, found = suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "monthly")
// 	suite.Require().True(found)
// 	suite.Require().Equal(epochInfo.CurrentEpoch, int64(0))
// 	suite.Require().Equal(epochInfo.CurrentEpochStartHeight, initialBlockHeight)
// 	suite.Require().Equal(epochInfo.CurrentEpochStartTime, time.Time{})
// 	suite.Require().Equal(epochInfo.EpochCountingStarted, false)
//
// 	// after 1 month
// 	suite.ctx = suite.ctx.WithBlockHeight(3).WithBlockTime(now.Add(month))
// 	suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
//
// 	// epoch started
// 	epochInfo, found = suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "monthly")
// 	suite.Require().True(found)
// 	suite.Require().Equal(epochInfo.CurrentEpoch, int64(1))
// 	suite.Require().Equal(epochInfo.CurrentEpochStartHeight, suite.ctx.BlockHeight())
// 	suite.Require().Equal(epochInfo.CurrentEpochStartTime.UTC().String(), now.Add(month).UTC().String())
// 	suite.Require().Equal(epochInfo.EpochCountingStarted, true)
// }
