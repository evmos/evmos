package keeper_test

import (
	"github.com/evmos/evmos/v10/x/inflation/types"
)

func (suite *KeeperTestSuite) TestParams() {
	testCases := []struct {
		name      string
		mockFunc  func() types.Params
		expParams types.Params
	}{
		{
			"Pass default params",
			func() types.Params {
				params := suite.app.InflationKeeper.GetParams(suite.ctx)
				return params
			},
			types.DefaultParams(),
		},
		{
			"pass - setting new params",
			func() types.Params {
				params := types.DefaultParams()
				err := suite.app.InflationKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				return params
			},
			suite.app.InflationKeeper.GetParams(suite.ctx),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			params := tc.mockFunc()
			suite.Require().Equal(tc.expParams, params)
		})
	}
}
