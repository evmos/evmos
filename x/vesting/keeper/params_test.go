package keeper_test

import "github.com/evmos/evmos/v14/x/vesting/types"

func (suite *KeeperTestSuite) TestParams() {
	testCases := []struct {
		name      string
		mockFunc  func() types.Params
		expParams types.Params
	}{
		{
			"Pass default params",
			func() types.Params {
				params := suite.app.VestingKeeper.GetParams(suite.ctx)
				return params
			},
			types.DefaultParams(),
		},
		{
			"pass - setting new params",
			func() types.Params {
				params := types.DefaultParams()
				err := suite.app.VestingKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				return params
			},
			suite.app.VestingKeeper.GetParams(suite.ctx),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			params := tc.mockFunc()
			suite.Require().Equal(tc.expParams, params)
		})
	}
}
