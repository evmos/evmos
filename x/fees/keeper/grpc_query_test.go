package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/v5/x/fees/types"
)

func (suite *KeeperTestSuite) TestDevFeeInfos() {
	var (
		req    *types.QueryDevFeeInfosRequest
		expRes *types.QueryDevFeeInfosResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"no fee infos registered",
			func() {
				req = &types.QueryDevFeeInfosRequest{}
				expRes = &types.QueryDevFeeInfosResponse{Pagination: &query.PageResponse{}}
			},
			true,
		},
		{
			"1 fee infos registered w/pagination",
			func() {
				req = &types.QueryDevFeeInfosRequest{
					Pagination: &query.PageRequest{Limit: 10, CountTotal: true},
				}
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)

				expRes = &types.QueryDevFeeInfosResponse{
					Pagination: &query.PageResponse{Total: 1},
					Fees: []types.DevFeeInfo{
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
				req = &types.QueryDevFeeInfosRequest{}
				contract2 := tests.GenerateAddress()
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)
				suite.app.FeesKeeper.SetFee(suite.ctx, contract2, deployer, nil)

				expRes = &types.QueryDevFeeInfosResponse{
					Pagination: &query.PageResponse{Total: 2},
					Fees: []types.DevFeeInfo{
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

			res, err := suite.queryClient.DevFeeInfos(ctx, req)
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

// Cases that cannot be tested in TestDevFeeInfos
func (suite *KeeperTestSuite) TestDevFeeInfosKeeper() {
	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)
	res, err := suite.app.FeesKeeper.DevFeeInfos(ctx, nil)
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

func (suite *KeeperTestSuite) TestDevFeeInfo() {
	var (
		req    *types.QueryDevFeeInfoRequest
		expRes *types.QueryDevFeeInfoResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"empty contract address",
			func() {
				req = &types.QueryDevFeeInfoRequest{}
				expRes = &types.QueryDevFeeInfoResponse{}
			},
			false,
		},
		{
			"invalid contract address",
			func() {
				req = &types.QueryDevFeeInfoRequest{
					ContractAddress: "1234",
				}
				expRes = &types.QueryDevFeeInfoResponse{}
			},
			false,
		},
		{
			"fee info not found",
			func() {
				req = &types.QueryDevFeeInfoRequest{
					ContractAddress: contract.String(),
				}
				expRes = &types.QueryDevFeeInfoResponse{}
			},
			false,
		},
		{
			"fee info found",
			func() {
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)

				req = &types.QueryDevFeeInfoRequest{
					ContractAddress: contract.Hex(),
				}
				expRes = &types.QueryDevFeeInfoResponse{Fee: types.DevFeeInfo{
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

			res, err := suite.queryClient.DevFeeInfo(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// Cases that cannot be tested in TestDevFeeInfo
func (suite *KeeperTestSuite) TestDevFeeInfoKeeper() {
	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)
	res, err := suite.app.FeesKeeper.DevFeeInfo(ctx, nil)
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

func (suite *KeeperTestSuite) TestDevFeeInfosPerDeployer() {
	var (
		req    *types.QueryDevFeeInfosPerDeployerRequest
		expRes *types.QueryDevFeeInfosPerDeployerResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"no contract registered",
			func() {
				req = &types.QueryDevFeeInfosPerDeployerRequest{}
				expRes = &types.QueryDevFeeInfosPerDeployerResponse{Pagination: &query.PageResponse{}}
			},
			false,
		},
		{
			"invalid deployer address",
			func() {
				req = &types.QueryDevFeeInfosPerDeployerRequest{
					DeployerAddress: "123",
				}
				expRes = &types.QueryDevFeeInfosPerDeployerResponse{Pagination: &query.PageResponse{}}
			},
			false,
		},
		{
			"1 fee info registered w/pagination",
			func() {
				req = &types.QueryDevFeeInfosPerDeployerRequest{
					Pagination:      &query.PageRequest{Limit: 10, CountTotal: true},
					DeployerAddress: deployer.String(),
				}
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)
				suite.app.FeesKeeper.SetFeeInverse(suite.ctx, deployer, contract)

				expRes = &types.QueryDevFeeInfosPerDeployerResponse{
					Pagination: &query.PageResponse{Total: 1},
					Fees: []types.DevFeeInfo{
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
				req = &types.QueryDevFeeInfosPerDeployerRequest{
					DeployerAddress: deployer.String(),
				}
				contract2 := tests.GenerateAddress()
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)
				suite.app.FeesKeeper.SetFeeInverse(suite.ctx, deployer, contract)
				suite.app.FeesKeeper.SetFee(suite.ctx, contract2, deployer, nil)
				suite.app.FeesKeeper.SetFeeInverse(suite.ctx, deployer, contract2)

				expRes = &types.QueryDevFeeInfosPerDeployerResponse{
					Pagination: &query.PageResponse{Total: 2},
					Fees: []types.DevFeeInfo{
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

			res, err := suite.queryClient.DevFeeInfosPerDeployer(ctx, req)
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

// Cases that cannot be tested in TestDevFeeInfosPerDeployer
func (suite *KeeperTestSuite) TestDevFeeInfosPerDeployerKeeper() {
	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)
	res, err := suite.app.FeesKeeper.DevFeeInfosPerDeployer(ctx, nil)
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
