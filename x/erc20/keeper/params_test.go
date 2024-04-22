package keeper_test

import (
	"reflect"

	"github.com/evmos/evmos/v17/x/erc20/types"
)

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.Erc20Keeper.GetParams(suite.ctx)
	suite.app.Erc20Keeper.SetParams(suite.ctx, params) //nolint:errcheck

	testCases := []struct {
		name      string
		paramsFun func() interface{}
		getFun    func() interface{}
		expected  bool
	}{
		{
			"success - Checks if the default params are set correctly",
			func() interface{} {
				return types.DefaultParams()
			},
			func() interface{} {
				return suite.app.Erc20Keeper.GetParams(suite.ctx)
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			outcome := reflect.DeepEqual(tc.paramsFun(), tc.getFun())
			suite.Require().Equal(tc.expected, outcome)
		})
	}
}
