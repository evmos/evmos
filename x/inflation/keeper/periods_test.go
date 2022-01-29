package keeper_test

import "fmt"

func (suite *KeeperTestSuite) TestSetGetPeriod() {
	expPeriod := uint64(9)

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
				suite.app.InflationKeeper.SetPeriod(suite.ctx, expPeriod)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			period := suite.app.InflationKeeper.GetPeriod(suite.ctx)
			if tc.ok {
				suite.Require().Equal(expPeriod, period, tc.name)
			} else {
				suite.Require().Zero(period, tc.name)
			}
		})
	}
}
