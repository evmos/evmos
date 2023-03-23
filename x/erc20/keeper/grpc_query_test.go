package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	utiltx "github.com/evmos/evmos/v12/testutil/tx"
	"github.com/evmos/evmos/v12/x/erc20/types"
)

func (suite *KeeperTestSuite) TestTokenPairs() {
	var (
		req    *types.QueryTokenPairsRequest
		expRes *types.QueryTokenPairsResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"no pairs registered",
			func() {
				req = &types.QueryTokenPairsRequest{}
				expRes = &types.QueryTokenPairsResponse{Pagination: &query.PageResponse{}}
			},
			true,
		},
		{
			"1 pair registered w/pagination",
			func() {
				req = &types.QueryTokenPairsRequest{
					Pagination: &query.PageRequest{Limit: 10, CountTotal: true},
				}
				pair := types.NewTokenPair(utiltx.GenerateAddress(), "coin", types.OWNER_MODULE)
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, pair)

				expRes = &types.QueryTokenPairsResponse{
					Pagination: &query.PageResponse{Total: 1},
					TokenPairs: []types.TokenPair{pair},
				}
			},
			true,
		},
		{
			"2 pairs registered wo/pagination",
			func() {
				req = &types.QueryTokenPairsRequest{}
				pair := types.NewTokenPair(utiltx.GenerateAddress(), "coin", types.OWNER_MODULE)
				pair2 := types.NewTokenPair(utiltx.GenerateAddress(), "coin2", types.OWNER_MODULE)
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, pair)
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, pair2)

				expRes = &types.QueryTokenPairsResponse{
					Pagination: &query.PageResponse{Total: 2},
					TokenPairs: []types.TokenPair{pair, pair2},
				}
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.malleate()

			res, err := suite.queryClient.TokenPairs(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.Pagination, res.Pagination)
				suite.Require().ElementsMatch(expRes.TokenPairs, res.TokenPairs)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestTokenPair() {
	var (
		req    *types.QueryTokenPairRequest
		expRes *types.QueryTokenPairResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"invalid token address",
			func() {
				req = &types.QueryTokenPairRequest{}
				expRes = &types.QueryTokenPairResponse{}
			},
			false,
		},
		{
			"token pair not found",
			func() {
				req = &types.QueryTokenPairRequest{
					Token: utiltx.GenerateAddress().Hex(),
				}
				expRes = &types.QueryTokenPairResponse{}
			},
			false,
		},
		{
			"token pair found",
			func() {
				addr := utiltx.GenerateAddress()
				pair := types.NewTokenPair(addr, "coin", types.OWNER_MODULE)
				suite.app.Erc20Keeper.SetTokenPair(suite.ctx, pair)
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, addr, pair.GetID())
				suite.app.Erc20Keeper.SetDenomMap(suite.ctx, pair.Denom, pair.GetID())

				req = &types.QueryTokenPairRequest{
					Token: pair.Erc20Address,
				}
				expRes = &types.QueryTokenPairResponse{TokenPair: pair}
			},
			true,
		},
		{
			"token pair not found - with erc20 existent",
			func() {
				addr := utiltx.GenerateAddress()
				pair := types.NewTokenPair(addr, "coin", types.OWNER_MODULE)
				suite.app.Erc20Keeper.SetERC20Map(suite.ctx, addr, pair.GetID())
				suite.app.Erc20Keeper.SetDenomMap(suite.ctx, pair.Denom, pair.GetID())

				req = &types.QueryTokenPairRequest{
					Token: pair.Erc20Address,
				}
				expRes = &types.QueryTokenPairResponse{TokenPair: pair}
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.malleate()

			res, err := suite.queryClient.TokenPair(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryParams() {
	ctx := sdk.WrapSDKContext(suite.ctx)
	expParams := types.DefaultParams()

	res, err := suite.queryClient.Params(ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expParams, res.Params)
}
