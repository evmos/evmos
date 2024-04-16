package keeper_test

import (
	"reflect"

	"github.com/evmos/evmos/v18/x/feemarket/types"
)

func (suite *KeeperTestSuite) TestGetParams() {
	params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
	suite.Require().NotNil(params.BaseFee)
	suite.Require().NotNil(params.MinGasPrice)
	suite.Require().NotNil(params.MinGasMultiplier)
}

func (suite *KeeperTestSuite) TestSetGetParams() {
	params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
	err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)
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
				return suite.app.FeeMarketKeeper.GetParams(suite.ctx)
			},
			true,
		},
		{
			"success - Check ElasticityMultiplier is set to 3 and can be retrieved correctly",
			func() interface{} {
				params.ElasticityMultiplier = 3
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				return params.ElasticityMultiplier
			},
			func() interface{} {
				return suite.app.FeeMarketKeeper.GetParams(suite.ctx).ElasticityMultiplier
			},
			true,
		},
		{
			"success - Check BaseFeeEnabled is computed with its default params and can be retrieved correctly",
			func() interface{} {
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, types.DefaultParams())
				suite.Require().NoError(err)
				return true
			},
			func() interface{} {
				return suite.app.FeeMarketKeeper.GetBaseFeeEnabled(suite.ctx)
			},
			true,
		},
		{
			"success - Check BaseFeeEnabled is computed with alternate params and can be retrieved correctly",
			func() interface{} {
				params.NoBaseFee = true
				params.EnableHeight = 5
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				return true
			},
			func() interface{} {
				return suite.app.FeeMarketKeeper.GetBaseFeeEnabled(suite.ctx)
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			outcome := reflect.DeepEqual(tc.paramsFun(), tc.getFun())
			suite.Require().Equal(tc.expected, outcome)
		})
	}
}
