package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	"github.com/evmos/evmos/v19/x/feemarket/types"
)

func (suite *KeeperTestSuite) TestQueryParams() {
	testCases := []struct {
		name    string
		expPass bool
	}{
		{
			"pass",
			true,
		},
	}
	for _, tc := range testCases {
		params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
		exp := &types.QueryParamsResponse{Params: params}

		res, err := suite.queryClient.Params(suite.ctx.Context(), &types.QueryParamsRequest{})
		if tc.expPass {
			suite.Require().Equal(exp, res, tc.name)
			suite.Require().NoError(err)
		} else {
			suite.Require().Error(err)
		}
	}
}

func (suite *KeeperTestSuite) TestQueryBaseFee() {
	var expRes *types.QueryBaseFeeResponse

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"pass - non-nil Base Fee",
			func() {
				baseFee := sdkmath.LegacyNewDecWithPrec(10, 2)
				suite.app.FeeMarketKeeper.SetBaseFee(suite.ctx, baseFee)

				expRes = &types.QueryBaseFeeResponse{BaseFee: &baseFee}
			},
			true,
		},
	}
	for _, tc := range testCases {
		tc.malleate()

		res, err := suite.queryClient.BaseFee(suite.ctx.Context(), &types.QueryBaseFeeRequest{})
		if tc.expPass {
			suite.Require().NotNil(res)
			suite.Require().Equal(expRes, res, tc.name)
			suite.Require().NoError(err)
		} else {
			suite.Require().Error(err)
		}
	}
}

func (suite *KeeperTestSuite) TestQueryBlockGas() {
	testCases := []struct {
		name    string
		expPass bool
	}{
		{
			"pass",
			true,
		},
	}
	for _, tc := range testCases {
		gas := suite.app.FeeMarketKeeper.GetBlockGasWanted(suite.ctx)
		exp := &types.QueryBlockGasResponse{Gas: int64(gas)} //#nosec G115

		res, err := suite.queryClient.BlockGas(suite.ctx.Context(), &types.QueryBlockGasRequest{})
		if tc.expPass {
			suite.Require().Equal(exp, res, tc.name)
			suite.Require().NoError(err)
		} else {
			suite.Require().Error(err)
		}
	}
}
