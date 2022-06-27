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
				fee := types.NewFee(contract, deployer, withdraw)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee)

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
				fee := types.NewFee(contract, deployer, withdraw)
				fee2 := types.NewFee(contract2, deployer, nil)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee2)

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
				fee := types.NewFee(contract, deployer, withdraw)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee)

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
			"1 fee registered w/pagination",
			func() {
				req = &types.QueryDeployerFeesRequest{
					Pagination:      &query.PageRequest{Limit: 10, CountTotal: true},
					DeployerAddress: deployer.String(),
				}

				fee := types.NewFee(contract, deployer, withdraw)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee)
				suite.app.FeesKeeper.SetDeployerMap(suite.ctx, deployer, contract)
				suite.app.FeesKeeper.SetWithdrawMap(suite.ctx, withdraw, contract)

				expRes = &types.QueryDeployerFeesResponse{
					Pagination: &query.PageResponse{Total: 1},
					ContractAddresses: []string{
						contract.Hex(),
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
				fee := types.NewFee(contract, deployer, withdraw)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee)
				suite.app.FeesKeeper.SetDeployerMap(suite.ctx, deployer, contract)
				suite.app.FeesKeeper.SetWithdrawMap(suite.ctx, withdraw, contract)

				fee2 := types.NewFee(contract2, deployer, nil)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee2)
				suite.app.FeesKeeper.SetDeployerMap(suite.ctx, deployer, contract2)

				expRes = &types.QueryDeployerFeesResponse{
					Pagination: &query.PageResponse{Total: 2},
					ContractAddresses: []string{
						contract.Hex(),
						contract2.Hex(),
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
				suite.Require().ElementsMatch(expRes.ContractAddresses, res.ContractAddresses)
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

func (suite *KeeperTestSuite) TestWithdrawFees() {
	var (
		req    *types.QueryWithdrawFeesRequest
		expRes *types.QueryWithdrawFeesResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"no contract registered",
			func() {
				req = &types.QueryWithdrawFeesRequest{}
				expRes = &types.QueryWithdrawFeesResponse{Pagination: &query.PageResponse{}}
			},
			false,
		},
		{
			"invalid withdraw address",
			func() {
				req = &types.QueryWithdrawFeesRequest{
					WithdrawAddress: "123",
				}
				expRes = &types.QueryWithdrawFeesResponse{Pagination: &query.PageResponse{}}
			},
			false,
		},
		{
			"1 fee registered w/pagination",
			func() {
				req = &types.QueryWithdrawFeesRequest{
					Pagination:      &query.PageRequest{Limit: 10, CountTotal: true},
					WithdrawAddress: withdraw.String(),
				}

				fee := types.NewFee(contract, deployer, withdraw)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee)
				suite.app.FeesKeeper.SetDeployerMap(suite.ctx, deployer, contract)
				suite.app.FeesKeeper.SetWithdrawMap(suite.ctx, withdraw, contract)

				expRes = &types.QueryWithdrawFeesResponse{
					Pagination: &query.PageResponse{Total: 1},
					ContractAddresses: []string{
						contract.Hex(),
					},
				}
			},
			true,
		},
		{
			"2 fees registered for one withdraw address wo/pagination",
			func() {
				req = &types.QueryWithdrawFeesRequest{
					WithdrawAddress: withdraw.String(),
				}
				contract2 := tests.GenerateAddress()
				deployer2 := sdk.AccAddress(tests.GenerateAddress().Bytes())

				fee := types.NewFee(contract, deployer, withdraw)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee)
				suite.app.FeesKeeper.SetDeployerMap(suite.ctx, deployer, contract)
				suite.app.FeesKeeper.SetWithdrawMap(suite.ctx, withdraw, contract)

				fee2 := types.NewFee(contract2, deployer2, withdraw)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee2)
				suite.app.FeesKeeper.SetDeployerMap(suite.ctx, deployer2, contract2)
				suite.app.FeesKeeper.SetWithdrawMap(suite.ctx, withdraw, contract2)

				expRes = &types.QueryWithdrawFeesResponse{
					Pagination: &query.PageResponse{Total: 2},
					ContractAddresses: []string{
						contract.Hex(),
						contract2.Hex(),
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

			res, err := suite.queryClient.WithdrawFees(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.Pagination, res.Pagination)
				suite.Require().ElementsMatch(expRes.ContractAddresses, res.ContractAddresses)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// Cases that cannot be tested in TestWithdrawFees
func (suite *KeeperTestSuite) TestWithdrawFeesKeeper() {
	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)
	res, err := suite.app.FeesKeeper.WithdrawFees(ctx, nil)
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
