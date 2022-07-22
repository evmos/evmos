package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/evmos/ethermint/tests"
	"github.com/evmos/evmos/v7/x/feesplit/types"
)

func (suite *KeeperTestSuite) TestFeeSplits() {
	var (
		req    *types.QueryFeeSplitsRequest
		expRes *types.QueryFeeSplitsResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"no fee infos registered",
			func() {
				req = &types.QueryFeeSplitsRequest{}
				expRes = &types.QueryFeeSplitsResponse{Pagination: &query.PageResponse{}}
			},
			true,
		},
		{
			"1 fee infos registered w/pagination",
			func() {
				req = &types.QueryFeeSplitsRequest{
					Pagination: &query.PageRequest{Limit: 10, CountTotal: true},
				}
				feeSplit := types.NewFeeSplit(contract, deployer, withdraw)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, feeSplit)

				expRes = &types.QueryFeeSplitsResponse{
					Pagination: &query.PageResponse{Total: 1},
					FeeSplits: []types.FeeSplit{
						{
							ContractAddress:   contract.Hex(),
							DeployerAddress:   deployer.String(),
							WithdrawerAddress: withdraw.String(),
						},
					},
				}
			},
			true,
		},
		{
			"2 fee infos registered wo/pagination",
			func() {
				req = &types.QueryFeeSplitsRequest{}
				contract2 := tests.GenerateAddress()
				feeSplit := types.NewFeeSplit(contract, deployer, withdraw)
				feeSplit2 := types.NewFeeSplit(contract2, deployer, nil)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, feeSplit)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, feeSplit2)

				expRes = &types.QueryFeeSplitsResponse{
					Pagination: &query.PageResponse{Total: 2},
					FeeSplits: []types.FeeSplit{
						{
							ContractAddress:   contract.Hex(),
							DeployerAddress:   deployer.String(),
							WithdrawerAddress: withdraw.String(),
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

			res, err := suite.queryClient.FeeSplits(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.Pagination, res.Pagination)
				suite.Require().ElementsMatch(expRes.FeeSplits, res.FeeSplits)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// Cases that cannot be tested in TestFees
func (suite *KeeperTestSuite) TestFeesplitKeeper() {
	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)
	res, err := suite.app.FeesplitKeeper.FeeSplits(ctx, nil)
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

func (suite *KeeperTestSuite) TestFee() {
	var (
		req    *types.QueryFeeSplitRequest
		expRes *types.QueryFeeSplitResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"empty contract address",
			func() {
				req = &types.QueryFeeSplitRequest{}
				expRes = &types.QueryFeeSplitResponse{}
			},
			false,
		},
		{
			"invalid contract address",
			func() {
				req = &types.QueryFeeSplitRequest{
					ContractAddress: "1234",
				}
				expRes = &types.QueryFeeSplitResponse{}
			},
			false,
		},
		{
			"fee info not found",
			func() {
				req = &types.QueryFeeSplitRequest{
					ContractAddress: contract.String(),
				}
				expRes = &types.QueryFeeSplitResponse{}
			},
			false,
		},
		{
			"fee info found",
			func() {
				feeSplit := types.NewFeeSplit(contract, deployer, withdraw)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, feeSplit)

				req = &types.QueryFeeSplitRequest{
					ContractAddress: contract.Hex(),
				}
				expRes = &types.QueryFeeSplitResponse{FeeSplit: types.FeeSplit{
					ContractAddress:   contract.Hex(),
					DeployerAddress:   deployer.String(),
					WithdrawerAddress: withdraw.String(),
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

			res, err := suite.queryClient.FeeSplit(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDeployerFees() {
	var (
		req    *types.QueryDeployerFeeSplitsRequest
		expRes *types.QueryDeployerFeeSplitsResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"no contract registered",
			func() {
				req = &types.QueryDeployerFeeSplitsRequest{}
				expRes = &types.QueryDeployerFeeSplitsResponse{Pagination: &query.PageResponse{}}
			},
			false,
		},
		{
			"invalid deployer address",
			func() {
				req = &types.QueryDeployerFeeSplitsRequest{
					DeployerAddress: "123",
				}
				expRes = &types.QueryDeployerFeeSplitsResponse{Pagination: &query.PageResponse{}}
			},
			false,
		},
		{
			"1 fee registered w/pagination",
			func() {
				req = &types.QueryDeployerFeeSplitsRequest{
					Pagination:      &query.PageRequest{Limit: 10, CountTotal: true},
					DeployerAddress: deployer.String(),
				}

				feeSplit := types.NewFeeSplit(contract, deployer, withdraw)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, feeSplit)
				suite.app.FeesplitKeeper.SetDeployerMap(suite.ctx, deployer, contract)
				suite.app.FeesplitKeeper.SetWithdrawerMap(suite.ctx, withdraw, contract)

				expRes = &types.QueryDeployerFeeSplitsResponse{
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
				req = &types.QueryDeployerFeeSplitsRequest{
					DeployerAddress: deployer.String(),
				}
				contract2 := tests.GenerateAddress()
				feeSplit := types.NewFeeSplit(contract, deployer, withdraw)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, feeSplit)
				suite.app.FeesplitKeeper.SetDeployerMap(suite.ctx, deployer, contract)
				suite.app.FeesplitKeeper.SetWithdrawerMap(suite.ctx, withdraw, contract)

				feeSplit2 := types.NewFeeSplit(contract2, deployer, nil)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, feeSplit2)
				suite.app.FeesplitKeeper.SetDeployerMap(suite.ctx, deployer, contract2)

				expRes = &types.QueryDeployerFeeSplitsResponse{
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

			res, err := suite.queryClient.DeployerFeeSplits(ctx, req)
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
func (suite *KeeperTestSuite) TestDeployerFeesplitKeeper() {
	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)
	res, err := suite.app.FeesplitKeeper.DeployerFeeSplits(ctx, nil)
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

func (suite *KeeperTestSuite) TestWithdrawerFeeSplits() {
	var (
		req    *types.QueryWithdrawerFeeSplitsRequest
		expRes *types.QueryWithdrawerFeeSplitsResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"no contract registered",
			func() {
				req = &types.QueryWithdrawerFeeSplitsRequest{}
				expRes = &types.QueryWithdrawerFeeSplitsResponse{Pagination: &query.PageResponse{}}
			},
			false,
		},
		{
			"invalid withdraw address",
			func() {
				req = &types.QueryWithdrawerFeeSplitsRequest{
					WithdrawerAddress: "123",
				}
				expRes = &types.QueryWithdrawerFeeSplitsResponse{Pagination: &query.PageResponse{}}
			},
			false,
		},
		{
			"1 fee registered w/pagination",
			func() {
				req = &types.QueryWithdrawerFeeSplitsRequest{
					Pagination:        &query.PageRequest{Limit: 10, CountTotal: true},
					WithdrawerAddress: withdraw.String(),
				}

				feeSplit := types.NewFeeSplit(contract, deployer, withdraw)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, feeSplit)
				suite.app.FeesplitKeeper.SetDeployerMap(suite.ctx, deployer, contract)
				suite.app.FeesplitKeeper.SetWithdrawerMap(suite.ctx, withdraw, contract)

				expRes = &types.QueryWithdrawerFeeSplitsResponse{
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
				req = &types.QueryWithdrawerFeeSplitsRequest{
					WithdrawerAddress: withdraw.String(),
				}
				contract2 := tests.GenerateAddress()
				deployer2 := sdk.AccAddress(tests.GenerateAddress().Bytes())

				feeSplit := types.NewFeeSplit(contract, deployer, withdraw)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, feeSplit)
				suite.app.FeesplitKeeper.SetDeployerMap(suite.ctx, deployer, contract)
				suite.app.FeesplitKeeper.SetWithdrawerMap(suite.ctx, withdraw, contract)

				feeSplit2 := types.NewFeeSplit(contract2, deployer2, withdraw)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, feeSplit2)
				suite.app.FeesplitKeeper.SetDeployerMap(suite.ctx, deployer2, contract2)
				suite.app.FeesplitKeeper.SetWithdrawerMap(suite.ctx, withdraw, contract2)

				expRes = &types.QueryWithdrawerFeeSplitsResponse{
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

			res, err := suite.queryClient.WithdrawerFeeSplits(ctx, req)
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

// Cases that cannot be tested in TestWithdrawerFees
func (suite *KeeperTestSuite) TestWithdrawerFeesplitKeeper() {
	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)
	res, err := suite.app.FeesplitKeeper.WithdrawerFeeSplits(ctx, nil)
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

func (suite *KeeperTestSuite) TestQueryParams() {
	ctx := sdk.WrapSDKContext(suite.ctx)
	expParams := types.DefaultParams()
	expParams.EnableFeeSplit = true

	res, err := suite.queryClient.Params(ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expParams, res.Params)
}
