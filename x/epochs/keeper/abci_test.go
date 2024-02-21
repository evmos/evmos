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

	now := time.Now()

	testCases := []struct {
		expCurrentEpochStartTime   time.Time
		expCurrentEpochStartHeight int64
		expCurrentEpoch            int64
		expInitialEpochStartTime   time.Time
		malleate                   func(ctx sdktypes.Context)
	}{
		{
			// Only advance 2 seconds, do not increment epoch
			expCurrentEpochStartHeight: 1,
			expCurrentEpochStartTime:   time.Time{},
			expCurrentEpoch:            1,
			expInitialEpochStartTime:   time.Time{},
			malleate: func(ctx sdktypes.Context) {
				epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(ctx, "month")
				fmt.Printf("%+v\n", epochInfo)
			},
		},
		{
			// Only advance 2 seconds, do not increment epoch
			expCurrentEpochStartHeight: 2,
			expCurrentEpochStartTime:   now,
			expCurrentEpoch:            2,
			expInitialEpochStartTime:   now,
			malleate: func(ctx sdktypes.Context) {
				ctx = ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				suite.network.App.EpochsKeeper.BeginBlocker(ctx)
				epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(ctx, "month")
				fmt.Printf("%+v\n", epochInfo)
			},
		},
		// // 	{
		// 		expCurrentEpochStartHeight: 2,
		// 		expCurrentEpochStartTime:   now,
		// 		expCurrentEpoch:            1,
		// 		expInitialEpochStartTime:   now,
		// 		fn: func() {
		// 			suite.ctx = suite.ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
		// 			suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
		// 			epochInfo, found = suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "monthly")
		// 			suite.Require().True(found)
		// 		},
		// 	},
		// 	{
		// 		expCurrentEpochStartHeight: 2,
		// 		expCurrentEpochStartTime:   now,
		// 		expCurrentEpoch:            1,
		// 		expInitialEpochStartTime:   now,
		// 		fn: func() {
		// 			suite.ctx = suite.ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
		// 			suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
		// 			suite.ctx = suite.ctx.WithBlockHeight(3).WithBlockTime(now.Add(time.Hour * 24 * 31))
		// 			suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
		// 			epochInfo, found = suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "monthly")
		// 			suite.Require().True(found)
		// 		},
		// 	},
		// 	// Test that incrementing _exactly_ 1 month increments the epoch count.
		// 	{
		// 		expCurrentEpochStartHeight: 3,
		// 		expCurrentEpochStartTime:   now.Add(time.Hour * 24 * 31),
		// 		expCurrentEpoch:            2,
		// 		expInitialEpochStartTime:   now,
		// 		fn: func() {
		// 			suite.ctx = suite.ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
		// 			suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
		// 			suite.ctx = suite.ctx.WithBlockHeight(3).WithBlockTime(now.Add(time.Hour * 24 * 32))
		// 			suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
		// 			epochInfo, found = suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "monthly")
		// 			suite.Require().True(found)
		// 		},
		// 	},
		// 	{ //nolint:dupl
		// 		expCurrentEpochStartHeight: 3,
		// 		expCurrentEpochStartTime:   now.Add(time.Hour * 24 * 31),
		// 		expCurrentEpoch:            2,
		// 		expInitialEpochStartTime:   now,
		// 		fn: func() {
		// 			suite.ctx = suite.ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
		// 			suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
		// 			suite.ctx = suite.ctx.WithBlockHeight(3).WithBlockTime(now.Add(time.Hour * 24 * 32))
		// 			suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
		// 			suite.ctx.WithBlockHeight(4).WithBlockTime(now.Add(time.Hour * 24 * 33))
		// 			suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
		// 			epochInfo, found = suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "monthly")
		// 			suite.Require().True(found)
		// 		},
		// 	},
		// 	{ //nolint:dupl
		// 		expCurrentEpochStartHeight: 3,
		// 		expCurrentEpochStartTime:   now.Add(time.Hour * 24 * 31),
		// 		expCurrentEpoch:            2,
		// 		expInitialEpochStartTime:   now,
		// 		fn: func() {
		// 			suite.ctx = suite.ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
		// 			suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
		// 			suite.ctx = suite.ctx.WithBlockHeight(3).WithBlockTime(now.Add(time.Hour * 24 * 32))
		// 			suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
		// 			suite.ctx.WithBlockHeight(4).WithBlockTime(now.Add(time.Hour * 24 * 33))
		// 			suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
		// 			epochInfo, found = suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "monthly")
		// 			suite.Require().True(found)
		// 		},
		// 	},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %d", i), func() {
			fmt.Println("TEST OF INTEREST")
			// custom genesis defines an epoch that is not yet start but that should start at
			// specific time and block. This should happen in the BeginBlocker.

			month := time.Hour * 24 * 31
			initialBlockHeight := int64(1)

			epochsInfo := []types.EpochInfo{
				{
					Identifier:              "month",
					StartTime:               time.Time{},
					Duration:                month,
					CurrentEpoch:            0,
					CurrentEpochStartHeight: initialBlockHeight,
					CurrentEpochStartTime:   time.Time{},
					EpochCountingStarted:    false,
				},
			}
			ctx = suite.SetupTest(epochsInfo) // reset

			// Check that custom genesis worked.
			epochInfo, found = suite.network.App.EpochsKeeper.GetEpochInfo(ctx, "monthly")
			suite.Require().True(found, "expected to find custom genesis data")

			ctx = ctx.WithBlockHeight(initialBlockHeight).WithBlockTime(now)
			tc.malleate(ctx)

			suite.Require().Equal("monthly", epochInfo.Identifier, "expected a different identifier")
			suite.Require().Equal(month, epochInfo.Duration, "expected a different duration")
			suite.Require().Equal(tc.expCurrentEpoch, epochInfo.CurrentEpoch, "expected a different current epoch")
			suite.Require().Equal(tc.expCurrentEpochStartHeight, epochInfo.CurrentEpochStartHeight, "expected different current epoch start height")
			suite.Require().Equal(tc.expCurrentEpochStartTime.UTC().String(), epochInfo.CurrentEpochStartTime.UTC().String(), "expected different current epoch start time")
			suite.Require().Equal(true, epochInfo.EpochCountingStarted, "expected different epoch counting started")
			suite.Require().Equal(tc.expInitialEpochStartTime.UTC().String(), epochInfo.StartTime.UTC().String(), "expected a different start time")
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
