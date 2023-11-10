package keeper_test

import (
	"github.com/evmos/evmos/v15/x/recovery/types"
)

func (suite *KeeperTestSuite) TestQueryParams() {
	ctx := suite.ctx
	expParams := types.DefaultParams()

	res, err := suite.queryClient.Params(ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expParams, res.Params)
}
