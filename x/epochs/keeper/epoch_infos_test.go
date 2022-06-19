package keeper_test

import (
	"time"

	"github.com/evmos/evmos/v5/x/epochs/types"
)

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
	suite.Require().Equal(allEpochs[0].Identifier, types.DayEpochID) // alphabetical order
	suite.Require().Equal(allEpochs[1].Identifier, "monthly")
	suite.Require().Equal(allEpochs[2].Identifier, types.WeekEpochID)
}
