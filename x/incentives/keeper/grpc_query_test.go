package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/tharsis/evmos/x/incentives/types"
)

func (suite *KeeperTestSuite) TestIncentives() {
	var (
		req    *types.QueryIncentivesRequest
		expRes *types.QueryIncentivesResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"no incentives registered",
			func() {
				req = &types.QueryIncentivesRequest{}
				expRes = &types.QueryIncentivesResponse{Pagination: &query.PageResponse{}}
			},
			true,
		},
		{
			"1 incentive registered w/pagination",
			func() {
				req = &types.QueryIncentivesRequest{
					Pagination: &query.PageRequest{Limit: 10, CountTotal: true},
				}
				in := types.NewIncentive(contract, allocations, epochs)
				suite.app.IncentivesKeeper.SetIncentive(suite.ctx, in)
				suite.Commit()

				expRes = &types.QueryIncentivesResponse{
					Pagination: &query.PageResponse{Total: 1},
					Incentives: []types.Incentive{in},
				}
			},
			true,
		},
		{
			"2 incentives registered wo/pagination",
			func() {
				req = &types.QueryIncentivesRequest{}
				in := types.NewIncentive(contract, allocations, epochs)
				in2 := types.NewIncentive(contract2, allocations, epochs)
				suite.app.IncentivesKeeper.SetIncentive(suite.ctx, in)
				suite.app.IncentivesKeeper.SetIncentive(suite.ctx, in2)
				suite.Commit()

				expRes = &types.QueryIncentivesResponse{
					Pagination: &query.PageResponse{Total: 2},
					Incentives: []types.Incentive{in, in2},
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

			res, err := suite.queryClient.Incentives(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.Pagination, res.Pagination)
				suite.Require().ElementsMatch(expRes.Incentives, res.Incentives)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestIncentive() {
	var (
		req    *types.QueryIncentiveRequest
		expRes *types.QueryIncentiveResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"invalid contract address",
			func() {
				req = &types.QueryIncentiveRequest{}
				expRes = &types.QueryIncentiveResponse{}
			},
			false,
		},
		{
			"gas meter not found",
			func() {
				req = &types.QueryIncentiveRequest{
					Contract: contract.String(),
				}
				expRes = &types.QueryIncentiveResponse{}
			},
			false,
		},
		{
			"gas meter found",
			func() {
				in := types.NewIncentive(contract, allocations, epochs)
				suite.app.IncentivesKeeper.SetIncentive(suite.ctx, in)
				suite.Commit()

				req = &types.QueryIncentiveRequest{
					Contract: contract.String(),
				}
				expRes = &types.QueryIncentiveResponse{Incentive: in}
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.malleate()

			res, err := suite.queryClient.Incentive(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGasMeters() {
	var (
		req    *types.QueryGasMetersRequest
		expRes *types.QueryGasMetersResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"no gas meter registered",
			func() {
				req = &types.QueryGasMetersRequest{}
				expRes = &types.QueryGasMetersResponse{Pagination: &query.PageResponse{}}
			},
			false,
		},
		{
			"1 gas meter registered w/pagination",
			func() {
				req = &types.QueryGasMetersRequest{
					Pagination: &query.PageRequest{Limit: 10, CountTotal: true},
					Contract:   contract.Hex(),
				}
				gm := types.NewGasMeter(contract, participant, 1)
				suite.app.IncentivesKeeper.SetGasMeter(suite.ctx, gm)
				suite.Commit()

				expRes = &types.QueryGasMetersResponse{
					Pagination: &query.PageResponse{Total: 1},
					GasMeters:  []types.GasMeter{gm},
				}
			},
			true,
		},
		{
			"2 gas meters registered for one contract wo/pagination",
			func() {
				req = &types.QueryGasMetersRequest{
					Contract: contract.Hex(),
				}
				gm := types.NewGasMeter(contract, participant, 1)
				gm2 := types.NewGasMeter(contract, participant2, 1)
				gm3 := types.NewGasMeter(contract2, participant, 1)
				suite.app.IncentivesKeeper.SetGasMeter(suite.ctx, gm)
				suite.app.IncentivesKeeper.SetGasMeter(suite.ctx, gm2)
				suite.app.IncentivesKeeper.SetGasMeter(suite.ctx, gm3)
				suite.Commit()

				expRes = &types.QueryGasMetersResponse{
					Pagination: &query.PageResponse{Total: 2},
					GasMeters:  []types.GasMeter{gm, gm2},
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

			res, err := suite.queryClient.GasMeters(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.Pagination, res.Pagination)
				suite.Require().ElementsMatch(expRes.GasMeters, res.GasMeters)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGasMeter() {
	var (
		req    *types.QueryGasMeterRequest
		expRes *types.QueryGasMeterResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"invalid token address",
			func() {
				req = &types.QueryGasMeterRequest{}
				expRes = &types.QueryGasMeterResponse{}
			},
			false,
		},
		{
			"invalid participant address",
			func() {
				req = &types.QueryGasMeterRequest{
					Contract:    contract.String(),
					Participant: "1234",
				}
				expRes = &types.QueryGasMeterResponse{}
			},
			false,
		},
		{
			"gas meter not found",
			func() {
				req = &types.QueryGasMeterRequest{
					Contract:    contract.String(),
					Participant: participant.String(),
				}
				expRes = &types.QueryGasMeterResponse{}
			},
			false,
		},
		{
			"gas meter found",
			func() {
				gm := types.NewGasMeter(contract, participant, 1)
				suite.app.IncentivesKeeper.SetGasMeter(suite.ctx, gm)
				suite.Commit()

				req = &types.QueryGasMeterRequest{
					Contract:    contract.String(),
					Participant: participant.String(),
				}
				expRes = &types.QueryGasMeterResponse{GasMeter: 1}
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.malleate()

			res, err := suite.queryClient.GasMeter(ctx, req)
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
