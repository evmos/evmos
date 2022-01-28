package keeper_test

import (
	"fmt"
	"time"
)

func (suite *KeeperTestSuite) TestPeriodChangesAfterEpochEnd() {
	suite.SetupTest()

	currentEpochPeriod := suite.app.InflationKeeper.GetEpochsPerPeriod(suite.ctx)

	testCases := []struct {
		name    string
		height  int64
		changes bool
	}{
		{
			"period stays the same under epoch per period",
			currentEpochPeriod - 10, // so it's within range
			false,
		},
		{
			"period changes once enough epochs have passed",
			currentEpochPeriod + 1,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			currentPeriod := suite.app.InflationKeeper.GetPeriod(suite.ctx)
			epochIdentifier := suite.app.InflationKeeper.GetEpochIdentifier(suite.ctx)
			futureCtx := suite.ctx.WithBlockTime(time.Now().Add(time.Minute))
			suite.app.EpochsKeeper.BeforeEpochStart(futureCtx, epochIdentifier, tc.height)
			suite.app.EpochsKeeper.AfterEpochEnd(futureCtx, epochIdentifier, tc.height)

			if tc.changes {
				newPeriod := currentPeriod + 1
				suite.Require().Equal(suite.app.InflationKeeper.GetPeriod(suite.ctx), newPeriod)
			} else {
				suite.Require().Equal(suite.app.InflationKeeper.GetPeriod(suite.ctx), currentPeriod)
			}
		})
	}
}
