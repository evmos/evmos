package keeper_test

import (
	"fmt"
	"time"

	"github.com/evmos/evmos/v16/x/epochs"
	"github.com/evmos/evmos/v16/x/epochs/types"
)

func (suite *KeeperTestSuite) TestEpochInfoChangesBeginBlockerAndInitGenesis() {
	var (
		epochInfo types.EpochInfo
		found     bool
	)

	now := time.Now()

	testCases := []struct {
		expCurrentEpochStartTime   time.Time
		expCurrentEpochStartHeight int64
		expCurrentEpoch            int64
		expInitialEpochStartTime   time.Time
		fn                         func()
	}{
		{
			// Only advance 2 seconds, do not increment epoch
			expCurrentEpochStartHeight: 2,
			expCurrentEpochStartTime:   now,
			expCurrentEpoch:            1,
			expInitialEpochStartTime:   now,
			fn: func() {
				suite.ctx = suite.ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
				epochInfo, found = suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "monthly")
				suite.Require().True(found)
			},
		},
		{
			expCurrentEpochStartHeight: 2,
			expCurrentEpochStartTime:   now,
			expCurrentEpoch:            1,
			expInitialEpochStartTime:   now,
			fn: func() {
				suite.ctx = suite.ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
				epochInfo, found = suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "monthly")
				suite.Require().True(found)
			},
		},
		{
			expCurrentEpochStartHeight: 2,
			expCurrentEpochStartTime:   now,
			expCurrentEpoch:            1,
			expInitialEpochStartTime:   now,
			fn: func() {
				suite.ctx = suite.ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
				suite.ctx = suite.ctx.WithBlockHeight(3).WithBlockTime(now.Add(time.Hour * 24 * 31))
				suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
				epochInfo, found = suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "monthly")
				suite.Require().True(found)
			},
		},
		// Test that incrementing _exactly_ 1 month increments the epoch count.
		{
			expCurrentEpochStartHeight: 3,
			expCurrentEpochStartTime:   now.Add(time.Hour * 24 * 31),
			expCurrentEpoch:            2,
			expInitialEpochStartTime:   now,
			fn: func() {
				suite.ctx = suite.ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
				suite.ctx = suite.ctx.WithBlockHeight(3).WithBlockTime(now.Add(time.Hour * 24 * 32))
				suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
				epochInfo, found = suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "monthly")
				suite.Require().True(found)
			},
		},
		{ //nolint:dupl
			expCurrentEpochStartHeight: 3,
			expCurrentEpochStartTime:   now.Add(time.Hour * 24 * 31),
			expCurrentEpoch:            2,
			expInitialEpochStartTime:   now,
			fn: func() {
				suite.ctx = suite.ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
				suite.ctx = suite.ctx.WithBlockHeight(3).WithBlockTime(now.Add(time.Hour * 24 * 32))
				suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
				suite.ctx.WithBlockHeight(4).WithBlockTime(now.Add(time.Hour * 24 * 33))
				suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
				epochInfo, found = suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "monthly")
				suite.Require().True(found)
			},
		},
		{ //nolint:dupl
			expCurrentEpochStartHeight: 3,
			expCurrentEpochStartTime:   now.Add(time.Hour * 24 * 31),
			expCurrentEpoch:            2,
			expInitialEpochStartTime:   now,
			fn: func() {
				suite.ctx = suite.ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
				suite.ctx = suite.ctx.WithBlockHeight(3).WithBlockTime(now.Add(time.Hour * 24 * 32))
				suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
				suite.ctx.WithBlockHeight(4).WithBlockTime(now.Add(time.Hour * 24 * 33))
				suite.app.EpochsKeeper.BeginBlocker(suite.ctx)
				epochInfo, found = suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "monthly")
				suite.Require().True(found)
			},
		},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %d", i), func() {
			suite.SetupTest() // reset

			// On init genesis, default epochs information is set
			// To check init genesis again, should make it fresh status
			epochInfos := suite.app.EpochsKeeper.AllEpochInfos(suite.ctx)
			for _, epochInfo := range epochInfos {
				suite.app.EpochsKeeper.DeleteEpochInfo(suite.ctx, epochInfo.Identifier)
			}

			suite.ctx = suite.ctx.WithBlockHeight(1).WithBlockTime(now)

			// check init genesis
			epochs.InitGenesis(suite.ctx, suite.app.EpochsKeeper, types.GenesisState{
				Epochs: []types.EpochInfo{
					{
						Identifier:              "monthly",
						StartTime:               time.Time{},
						Duration:                time.Hour * 24 * 31,
						CurrentEpoch:            0,
						CurrentEpochStartHeight: suite.ctx.BlockHeight(),
						CurrentEpochStartTime:   time.Time{},
						EpochCountingStarted:    false,
					},
				},
			})

			tc.fn()

			suite.Require().Equal(epochInfo.Identifier, "monthly")
			suite.Require().Equal(epochInfo.StartTime.UTC().String(), tc.expInitialEpochStartTime.UTC().String())
			suite.Require().Equal(epochInfo.Duration, time.Hour*24*31)
			suite.Require().Equal(epochInfo.CurrentEpoch, tc.expCurrentEpoch)
			suite.Require().Equal(epochInfo.CurrentEpochStartHeight, tc.expCurrentEpochStartHeight)
			suite.Require().Equal(epochInfo.CurrentEpochStartTime.UTC().String(), tc.expCurrentEpochStartTime.UTC().String())
			suite.Require().Equal(epochInfo.EpochCountingStarted, true)
		})
	}
}

func (suite *KeeperTestSuite) TestEpochStartingOneMonthAfterInitGenesis() {
	// On init genesis, default epochs information is set
	// To check init genesis again, should make it fresh status
	epochInfos := suite.app.EpochsKeeper.AllEpochInfos(suite.ctx)
	for _, epochInfo := range epochInfos {
		suite.app.EpochsKeeper.DeleteEpochInfo(suite.ctx, epochInfo.Identifier)
	}

	now := time.Now()
	week := time.Hour * 24 * 7
	month := time.Hour * 24 * 30
	initialBlockHeight := int64(1)
	suite.ctx = suite.ctx.WithBlockHeight(initialBlockHeight).WithBlockTime(now)

	epochs.InitGenesis(suite.ctx, suite.app.EpochsKeeper, types.GenesisState{
		Epochs: []types.EpochInfo{
			{
				Identifier:              "monthly",
				StartTime:               now.Add(month),
				Duration:                time.Hour * 24 * 30,
				CurrentEpoch:            0,
				CurrentEpochStartHeight: suite.ctx.BlockHeight(),
				CurrentEpochStartTime:   time.Time{},
				EpochCountingStarted:    false,
			},
		},
	})

	// epoch not started yet
	epochInfo, found := suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "monthly")
	suite.Require().True(found)
	suite.Require().Equal(epochInfo.CurrentEpoch, int64(0))
	suite.Require().Equal(epochInfo.CurrentEpochStartHeight, initialBlockHeight)
	suite.Require().Equal(epochInfo.CurrentEpochStartTime, time.Time{})
	suite.Require().Equal(epochInfo.EpochCountingStarted, false)

	// after 1 week
	suite.ctx = suite.ctx.WithBlockHeight(2).WithBlockTime(now.Add(week))
	suite.app.EpochsKeeper.BeginBlocker(suite.ctx)

	// epoch not started yet
	epochInfo, found = suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "monthly")
	suite.Require().True(found)
	suite.Require().Equal(epochInfo.CurrentEpoch, int64(0))
	suite.Require().Equal(epochInfo.CurrentEpochStartHeight, initialBlockHeight)
	suite.Require().Equal(epochInfo.CurrentEpochStartTime, time.Time{})
	suite.Require().Equal(epochInfo.EpochCountingStarted, false)

	// after 1 month
	suite.ctx = suite.ctx.WithBlockHeight(3).WithBlockTime(now.Add(month))
	suite.app.EpochsKeeper.BeginBlocker(suite.ctx)

	// epoch started
	epochInfo, found = suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "monthly")
	suite.Require().True(found)
	suite.Require().Equal(epochInfo.CurrentEpoch, int64(1))
	suite.Require().Equal(epochInfo.CurrentEpochStartHeight, suite.ctx.BlockHeight())
	suite.Require().Equal(epochInfo.CurrentEpochStartTime.UTC().String(), now.Add(month).UTC().String())
	suite.Require().Equal(epochInfo.EpochCountingStarted, true)
}
