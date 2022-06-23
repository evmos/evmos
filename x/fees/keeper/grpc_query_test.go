package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/evmos/evmos/v5/x/fees/types"
	"github.com/tharsis/ethermint/tests"
)

func (suite *KeeperTestSuite) TestFees() {
	var (
		req    *types.QueryFeesRequest
		expRes *types.QueryFeesResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"no fee infos registered",
			func() {
				req = &types.QueryFeesRequest{}
				expRes = &types.QueryFeesResponse{Pagination: &query.PageResponse{}}
			},
			true,
		},
		{
			"1 fee infos registered w/pagination",
			func() {
				req = &types.QueryFeesRequest{
					Pagination: &query.PageRequest{Limit: 10, CountTotal: true},
				}
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)

				expRes = &types.QueryFeesResponse{
					Pagination: &query.PageResponse{Total: 1},
					Fees: []types.Fee{
						{
							ContractAddress: contract.Hex(),
							DeployerAddress: deployer.String(),
							WithdrawAddress: withdraw.String(),
						},
					},
				}
			},
			true,
		},
		{
			"2 fee infos registered wo/pagination",
			func() {
				req = &types.QueryFeesRequest{}
				contract2 := tests.GenerateAddress()
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)
				suite.app.FeesKeeper.SetFee(suite.ctx, contract2, deployer, nil)

				expRes = &types.QueryFeesResponse{
					Pagination: &query.PageResponse{Total: 2},
					Fees: []types.Fee{
						{
							ContractAddress: contract.Hex(),
							DeployerAddress: deployer.String(),
							WithdrawAddress: withdraw.String(),
						},
						{
							ContractAddress: contract2.Hex(),
							DeployerAddress: deployer.String(),
						},
					},
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

			res, err := suite.queryClient.Fees(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.Pagination, res.Pagination)
				suite.Require().ElementsMatch(expRes.Fees, res.Fees)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// Cases that cannot be tested in TestFees
func (suite *KeeperTestSuite) TestFeesKeeper() {
	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)
	res, err := suite.app.FeesKeeper.Fees(ctx, nil)
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

func (suite *KeeperTestSuite) TestFee() {
	var (
		req    *types.QueryFeeRequest
		expRes *types.QueryFeeResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"empty contract address",
			func() {
				req = &types.QueryFeeRequest{}
				expRes = &types.QueryFeeResponse{}
			},
			false,
		},
		{
			"invalid contract address",
			func() {
				req = &types.QueryFeeRequest{
					ContractAddress: "1234",
				}
				expRes = &types.QueryFeeResponse{}
			},
			false,
		},
		{
			"fee info not found",
			func() {
				req = &types.QueryFeeRequest{
					ContractAddress: contract.String(),
				}
				expRes = &types.QueryFeeResponse{}
			},
			false,
		},
		{
			"fee info found",
			func() {
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)

				req = &types.QueryFeeRequest{
					ContractAddress: contract.Hex(),
				}
				expRes = &types.QueryFeeResponse{Fee: types.Fee{
					ContractAddress: contract.Hex(),
					DeployerAddress: deployer.String(),
					WithdrawAddress: withdraw.String(),
				}}
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.malleate()

			res, err := suite.queryClient.Fee(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// Cases that cannot be tested in TestFee
func (suite *KeeperTestSuite) TestFeeKeeper() {
	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)
	res, err := suite.app.FeesKeeper.Fee(ctx, nil)
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

func (suite *KeeperTestSuite) TestDeployerFees() {
	var (
		req    *types.QueryDeployerFeesRequest
		expRes *types.QueryDeployerFeesResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"no contract registered",
			func() {
				req = &types.QueryDeployerFeesRequest{}
				expRes = &types.QueryDeployerFeesResponse{Pagination: &query.PageResponse{}}
			},
			false,
		},
		{
			"invalid deployer address",
			func() {
				req = &types.QueryDeployerFeesRequest{
					DeployerAddress: "123",
				}
				expRes = &types.QueryDeployerFeesResponse{Pagination: &query.PageResponse{}}
			},
			false,
		},
		{
			"1 fee info registered w/pagination",
			func() {
				req = &types.QueryDeployerFeesRequest{
					Pagination:      &query.PageRequest{Limit: 10, CountTotal: true},
					DeployerAddress: deployer.String(),
				}
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)
				suite.app.FeesKeeper.SetDeployerFees(suite.ctx, deployer, contract)

				expRes = &types.QueryDeployerFeesResponse{
					Pagination: &query.PageResponse{Total: 1},
					Fees: []types.Fee{
						{
							ContractAddress: contract.Hex(),
							DeployerAddress: deployer.String(),
							WithdrawAddress: withdraw.String(),
						},
					},
				}
			},
			true,
		},
		{
			"2 fee infos registered for one contract wo/pagination",
			func() {
				req = &types.QueryDeployerFeesRequest{
					DeployerAddress: deployer.String(),
				}
				contract2 := tests.GenerateAddress()
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)
				suite.app.FeesKeeper.SetDeployerFees(suite.ctx, deployer, contract)
				suite.app.FeesKeeper.SetFee(suite.ctx, contract2, deployer, nil)
				suite.app.FeesKeeper.SetDeployerFees(suite.ctx, deployer, contract2)

				expRes = &types.QueryDeployerFeesResponse{
					Pagination: &query.PageResponse{Total: 2},
					Fees: []types.Fee{
						{
							ContractAddress: contract.Hex(),
							DeployerAddress: deployer.String(),
							WithdrawAddress: withdraw.String(),
						},
						{
							ContractAddress: contract2.Hex(),
							DeployerAddress: deployer.String(),
						},
					},
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

			res, err := suite.queryClient.DeployerFees(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.Pagination, res.Pagination)
				suite.Require().ElementsMatch(expRes.Fees, res.Fees)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// Cases that cannot be tested in TestDeployerFees
func (suite *KeeperTestSuite) TestDeployerFeesKeeper() {
	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)
	res, err := suite.app.FeesKeeper.DeployerFees(ctx, nil)
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

func (suite *KeeperTestSuite) TestQueryParams() {
	ctx := sdk.WrapSDKContext(suite.ctx)
	expParams := types.DefaultParams()
	expParams.EnableFees = true

	res, err := suite.queryClient.Params(ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expParams, res.Params)
}
