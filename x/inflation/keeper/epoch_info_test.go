package keeper_test

import (
	"fmt"

	"github.com/tharsis/evmos/x/inflation/types"
)

func (suite *KeeperTestSuite) TestSetGetEpochIdentifier() {
	defaultEpochIdentifier := types.DefaultGenesisState().EpochIdentifier
	expEpochIdentifier := "week"

	testCases := []struct {
		name     string
		malleate func()
		ok       bool
	}{
		{
			"default epochIdentifier",
			func() {},
			false,
		},
		{
			"epochIdentifier set",
			func() {
				suite.app.InflationKeeper.SetEpochIdentifier(suite.ctx, expEpochIdentifier)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			epochIdentifier := suite.app.InflationKeeper.GetEpochIdentifier(suite.ctx)
			if tc.ok {
				suite.Require().Equal(expEpochIdentifier, epochIdentifier, tc.name)
			} else {
				suite.Require().Equal(defaultEpochIdentifier, epochIdentifier, tc.name)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSetGetEpochsPerPeriod() {
	defaultEpochsPerPeriod := types.DefaultGenesisState().EpochsPerPeriod
	expEpochsPerPeriod := int64(180)

	testCases := []struct {
		name     string
		malleate func()
		ok       bool
	}{
		{
			"default period",
			func() {},
			false,
		},
		{
			"period set",
			func() {
				suite.app.InflationKeeper.SetEpochsPerPeriod(suite.ctx, expEpochsPerPeriod)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			epochsPerPeriod := suite.app.InflationKeeper.GetEpochsPerPeriod(suite.ctx)
			if tc.ok {
				suite.Require().Equal(expEpochsPerPeriod, epochsPerPeriod, tc.name)
			} else {
				suite.Require().Equal(defaultEpochsPerPeriod, epochsPerPeriod, tc.name)
			}
		})
	}
}
