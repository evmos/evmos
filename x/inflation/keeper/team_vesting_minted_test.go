package keeper_test

import (
	"fmt"
)

// TODO
func (suite *KeeperTestSuite) TestSetGetTeamVestingMinted() {
	testCases := []struct {
		name                 string
		malleate             func()
		expTeamVestingMinted bool
	}{
		{
			"after genesis",
			func() {},
			true,
		},
		{
			"set to false",
			func() {
				suite.app.InflationKeeper.SetTeamVestingMinted(suite.ctx, false)
			},
			false,
		},
		{
			"set to true",
			func() {
				suite.app.InflationKeeper.SetTeamVestingMinted(suite.ctx, false)
				suite.app.InflationKeeper.SetTeamVestingMinted(suite.ctx, true)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			teamVestingMinted := suite.app.InflationKeeper.IsTeamVestingMinted(suite.ctx)
			if tc.expTeamVestingMinted {
				suite.Require().True(teamVestingMinted, tc.name)
			} else {
				suite.Require().False(teamVestingMinted, tc.name)
			}
		})
	}
}
