package keeper_test

import (
	"sort"
	"time"

	"github.com/tharsis/evmos/v4/x/epochs/types"
)

type sortEpochInfos []types.EpochInfo

func (s sortEpochInfos) Len() int { return len(s) }
func (s sortEpochInfos) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s sortEpochInfos) Less(i, j int) bool {
	return s[i].Duration < s[j].Duration
}

func (suite *KeeperTestSuite) TestEpochLifeCycle() {
	suite.SetupTest()

	epochInfo := types.EpochInfo{
		Identifier:            "monthly",
		StartTime:             time.Time{},
		Duration:              time.Hour * 24 * 30,
		CurrentEpoch:          0,
		CurrentEpochStartTime: time.Time{},
		EpochCountingStarted:  false,
	}
	suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochInfo)
	epochInfoSaved, found := suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, "monthly")
	suite.Require().True(found)
	suite.Require().Equal(epochInfo, epochInfoSaved)

	allEpochs := suite.app.EpochsKeeper.AllEpochInfos(suite.ctx)
	suite.Require().Len(allEpochs, 3)

	// ascending numerical order
	suite.Require().Equal(allEpochs[0].Identifier, types.DayEpochID)
	suite.Require().Equal(allEpochs[1].Identifier, types.WeekEpochID)
	suite.Require().Equal(allEpochs[2].Identifier, "monthly")
}

func (suite *KeeperTestSuite) TestIterateEpochInfo() {
	suite.SetupTest()

	epochInfos := sortEpochInfos{
		{
			Identifier:              "day",
			StartTime:               time.Time{},
			Duration:                time.Hour * 24,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
		{
			Identifier:              "hour",
			StartTime:               time.Time{},
			Duration:                time.Hour,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
		{
			Identifier:            "monthly",
			StartTime:             time.Time{},
			Duration:              time.Hour * 24 * 30,
			CurrentEpoch:          0,
			CurrentEpochStartTime: time.Time{},
			EpochCountingStarted:  false,
		},
		{
			Identifier:              "week",
			StartTime:               time.Time{},
			Duration:                time.Hour * 24 * 7,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
	}

	for _, epochInfo := range epochInfos {
		suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochInfo)
	}

	sort.Sort(epochInfos)
	suite.app.EpochsKeeper.IterateEpochInfo(suite.ctx, func(index int64, epochInfo types.EpochInfo) bool {
		expectedEpoch := epochInfos[index]
		suite.Require().Equal(expectedEpoch.Identifier, epochInfo.Identifier)
		suite.Require().Equal(expectedEpoch.Duration, epochInfo.Duration)
		return false
	})
}

func (suite *KeeperTestSuite) TestAllEpochInfos() {
	suite.SetupTest()

	epochInfos := sortEpochInfos{
		{
			Identifier:              "day",
			StartTime:               time.Time{},
			Duration:                time.Hour * 24,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
		{
			Identifier:              "hour",
			StartTime:               time.Time{},
			Duration:                time.Hour,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
		{
			Identifier:            "monthly",
			StartTime:             time.Time{},
			Duration:              time.Hour * 24 * 30,
			CurrentEpoch:          0,
			CurrentEpochStartTime: time.Time{},
			EpochCountingStarted:  false,
		},
		{
			Identifier:              "week",
			StartTime:               time.Time{},
			Duration:                time.Hour * 24 * 7,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
	}

	for _, epochInfo := range epochInfos {
		suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochInfo)
	}

	// sorts epochs by ascending duration
	sort.Sort(epochInfos)
	storedEpochInfos := suite.app.EpochsKeeper.AllEpochInfos(suite.ctx)
	for i, epochInfo := range storedEpochInfos {
		expectedEpoch := epochInfos[i]
		suite.Require().Equal(expectedEpoch.Identifier, epochInfo.Identifier)
		suite.Require().Equal(expectedEpoch.Duration, epochInfo.Duration)
	}
}
