package keeper_test

import (
	"time"

	"github.com/evmos/evmos/v16/x/epochs/types"
)

func (suite *KeeperTestSuite) TestEpochLifeCycle() {
    // The default genesis includes day and week epochs.
	suite.SetupTest([]types.EpochInfo{})

	epochInfo := types.EpochInfo{
		Identifier:            "month",
		StartTime:             time.Time{},
		Duration:              time.Hour * 24 * 30,
		CurrentEpoch:          0,
		CurrentEpochStartTime: time.Time{},
		EpochCountingStarted:  false,
	}
    ctx := suite.network.GetContext()
	suite.network.App.EpochsKeeper.SetEpochInfo(ctx, epochInfo)
	epochInfoSaved, found := suite.network.App.EpochsKeeper.GetEpochInfo(ctx, "month")
	suite.Require().True(found)
	suite.Require().Equal(epochInfo, epochInfoSaved)

	allEpochs := suite.network.App.EpochsKeeper.AllEpochInfos(ctx)
	suite.Require().Len(allEpochs, 3)
	suite.Require().Equal(allEpochs[0].Identifier, types.DayEpochID) // alphabetical order
	suite.Require().Equal(allEpochs[1].Identifier, "month")
	suite.Require().Equal(allEpochs[2].Identifier, types.WeekEpochID)
}
