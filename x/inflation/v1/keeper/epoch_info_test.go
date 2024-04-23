package keeper_test

import (
	"fmt"

	epochstypes "github.com/evmos/evmos/v18/x/epochs/types"
	"github.com/evmos/evmos/v18/x/inflation/v1/types"
)

func (suite *KeeperTestSuite) TestSetGetEpochIdentifier() {
	defaultEpochIdentifier := types.DefaultGenesisState().EpochIdentifier
	expEpochIdentifier := epochstypes.WeekEpochID

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

func (suite *KeeperTestSuite) TestSetGetSkippedEpochs() {
	defaultSkippedEpochs := types.DefaultGenesisState().SkippedEpochs
	expSkippedepochs := uint64(20)

	testCases := []struct {
		name     string
		malleate func()
		ok       bool
	}{
		{
			"default skipped epoch",
			func() {},
			false,
		},
		{
			"skipped epoch set",
			func() {
				suite.app.InflationKeeper.SetSkippedEpochs(suite.ctx, expSkippedepochs)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			epochsPerPeriod := suite.app.InflationKeeper.GetSkippedEpochs(suite.ctx)
			if tc.ok {
				suite.Require().Equal(expSkippedepochs, epochsPerPeriod, tc.name)
			} else {
				suite.Require().Equal(defaultSkippedEpochs, epochsPerPeriod, tc.name)
			}
		})
	}
}
