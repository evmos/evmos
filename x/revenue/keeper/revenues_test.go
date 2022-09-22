package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/tests"
	"github.com/evmos/evmos/v8/x/revenue/types"
)

func (suite *KeeperTestSuite) TestGetFees() {
	var expRes []types.Revenue

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"no fee splits registered",
			func() { expRes = []types.Revenue{} },
		},
		{
			"one fee split registered with withdraw address",
			func() {
				feeSplit := types.NewRevenue(contract, deployer, withdraw)
				suite.app.RevenueKeeper.SetRevenue(suite.ctx, feeSplit)
				expRes = []types.Revenue{feeSplit}
			},
		},
		{
			"one fee split registered with no withdraw address",
			func() {
				feeSplit := types.NewRevenue(contract, deployer, nil)
				suite.app.RevenueKeeper.SetRevenue(suite.ctx, feeSplit)
				expRes = []types.Revenue{feeSplit}
			},
		},
		{
			"multiple fee splits registered",
			func() {
				deployer2 := sdk.AccAddress(tests.GenerateAddress().Bytes())
				contract2 := tests.GenerateAddress()
				contract3 := tests.GenerateAddress()
				feeSplit := types.NewRevenue(contract, deployer, withdraw)
				feeSplit2 := types.NewRevenue(contract2, deployer, nil)
				feeSplit3 := types.NewRevenue(contract3, deployer2, nil)
				suite.app.RevenueKeeper.SetRevenue(suite.ctx, feeSplit)
				suite.app.RevenueKeeper.SetRevenue(suite.ctx, feeSplit2)
				suite.app.RevenueKeeper.SetRevenue(suite.ctx, feeSplit3)
				expRes = []types.Revenue{feeSplit, feeSplit2, feeSplit3}
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			tc.malleate()

			res := suite.app.RevenueKeeper.GetRevenues(suite.ctx)
			suite.Require().ElementsMatch(expRes, res, tc.name)
		})
	}
}

func (suite *KeeperTestSuite) TestIterateFees() {
	var expRes []types.Revenue

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"no fee splits registered",
			func() { expRes = []types.Revenue{} },
		},
		{
			"one fee split registered with withdraw address",
			func() {
				feeSplit := types.NewRevenue(contract, deployer, withdraw)
				suite.app.RevenueKeeper.SetRevenue(suite.ctx, feeSplit)
				expRes = []types.Revenue{
					types.NewRevenue(contract, deployer, withdraw),
				}
			},
		},
		{
			"one fee split registered with no withdraw address",
			func() {
				feeSplit := types.NewRevenue(contract, deployer, nil)
				suite.app.RevenueKeeper.SetRevenue(suite.ctx, feeSplit)
				expRes = []types.Revenue{
					types.NewRevenue(contract, deployer, nil),
				}
			},
		},
		{
			"multiple fee splits registered",
			func() {
				deployer2 := sdk.AccAddress(tests.GenerateAddress().Bytes())
				contract2 := tests.GenerateAddress()
				contract3 := tests.GenerateAddress()
				feeSplit := types.NewRevenue(contract, deployer, withdraw)
				feeSplit2 := types.NewRevenue(contract2, deployer, nil)
				feeSplit3 := types.NewRevenue(contract3, deployer2, nil)
				suite.app.RevenueKeeper.SetRevenue(suite.ctx, feeSplit)
				suite.app.RevenueKeeper.SetRevenue(suite.ctx, feeSplit2)
				suite.app.RevenueKeeper.SetRevenue(suite.ctx, feeSplit3)
				expRes = []types.Revenue{feeSplit, feeSplit2, feeSplit3}
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			tc.malleate()

			suite.app.RevenueKeeper.IterateRevenues(suite.ctx, func(feeSplit types.Revenue) (stop bool) {
				suite.Require().Contains(expRes, feeSplit, tc.name)
				return false
			})
		})
	}
}

func (suite *KeeperTestSuite) TestGetRevenue() {
	testCases := []struct {
		name        string
		contract    common.Address
		deployer    sdk.AccAddress
		withdraw    sdk.AccAddress
		found       bool
		expWithdraw bool
	}{
		{
			"fee with no withdraw address",
			contract,
			deployer,
			nil,
			true,
			false,
		},
		{
			"fee with withdraw address same as deployer",
			contract,
			deployer,
			deployer,
			true,
			false,
		},
		{
			"fee with withdraw address same as contract",
			contract,
			deployer,
			sdk.AccAddress(contract.Bytes()),
			true,
			true,
		},
		{
			"fee with withdraw address different than deployer",
			contract,
			deployer,
			withdraw,
			true,
			true,
		},
		{
			"no fee",
			common.Address{},
			nil,
			nil,
			false,
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			if tc.found {
				feeSplit := types.NewRevenue(tc.contract, tc.deployer, tc.withdraw)
				if tc.deployer.Equals(tc.withdraw) {
					feeSplit.WithdrawerAddress = ""
				}

				suite.app.RevenueKeeper.SetRevenue(suite.ctx, feeSplit)
				suite.app.RevenueKeeper.SetDeployerMap(suite.ctx, tc.deployer, tc.contract)
			}

			if tc.expWithdraw {
				suite.app.RevenueKeeper.SetWithdrawerMap(suite.ctx, tc.withdraw, tc.contract)
			}

			feeSplit, found := suite.app.RevenueKeeper.GetRevenue(suite.ctx, tc.contract)
			foundD := suite.app.RevenueKeeper.IsDeployerMapSet(suite.ctx, tc.deployer, tc.contract)
			foundW := suite.app.RevenueKeeper.IsWithdrawerMapSet(suite.ctx, tc.withdraw, tc.contract)

			if tc.found {
				suite.Require().True(found, tc.name)
				suite.Require().Equal(tc.deployer.String(), feeSplit.DeployerAddress, tc.name)
				suite.Require().Equal(tc.contract.Hex(), feeSplit.ContractAddress, tc.name)

				suite.Require().True(foundD, tc.name)

				if tc.expWithdraw {
					suite.Require().Equal(tc.withdraw.String(), feeSplit.WithdrawerAddress, tc.name)
					suite.Require().True(foundW, tc.name)
				} else {
					suite.Require().Equal("", feeSplit.WithdrawerAddress, tc.name)
					suite.Require().False(foundW, tc.name)
				}
			} else {
				suite.Require().False(found, tc.name)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDeleteRevenue() {
	feeSplit := types.NewRevenue(contract, deployer, withdraw)
	suite.app.RevenueKeeper.SetRevenue(suite.ctx, feeSplit)

	initialFee, found := suite.app.RevenueKeeper.GetRevenue(suite.ctx, contract)
	suite.Require().True(found)

	testCases := []struct {
		name     string
		malleate func()
		ok       bool
	}{
		{"existing fee split", func() {}, true},
		{
			"deleted fee split",
			func() {
				suite.app.RevenueKeeper.DeleteRevenue(suite.ctx, feeSplit)
			},
			false,
		},
	}
	for _, tc := range testCases {
		tc.malleate()
		feeSplit, found := suite.app.RevenueKeeper.GetRevenue(suite.ctx, contract)
		if tc.ok {
			suite.Require().True(found, tc.name)
			suite.Require().Equal(initialFee, feeSplit, tc.name)
		} else {
			suite.Require().False(found, tc.name)
			suite.Require().Equal(types.Revenue{}, feeSplit, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestDeleteDeployerMap() {
	suite.app.RevenueKeeper.SetDeployerMap(suite.ctx, deployer, contract)
	found := suite.app.RevenueKeeper.IsDeployerMapSet(suite.ctx, deployer, contract)
	suite.Require().True(found)

	testCases := []struct {
		name     string
		malleate func()
		ok       bool
	}{
		{"existing deployer", func() {}, true},
		{
			"deleted deployer",
			func() {
				suite.app.RevenueKeeper.DeleteDeployerMap(suite.ctx, deployer, contract)
			},
			false,
		},
	}
	for _, tc := range testCases {
		tc.malleate()
		found := suite.app.RevenueKeeper.IsDeployerMapSet(suite.ctx, deployer, contract)
		if tc.ok {
			suite.Require().True(found, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestDeleteWithdrawMap() {
	suite.app.RevenueKeeper.SetWithdrawerMap(suite.ctx, withdraw, contract)
	found := suite.app.RevenueKeeper.IsWithdrawerMapSet(suite.ctx, withdraw, contract)
	suite.Require().True(found)

	testCases := []struct {
		name     string
		malleate func()
		ok       bool
	}{
		{"existing withdraw", func() {}, true},
		{
			"deleted withdraw",
			func() {
				suite.app.RevenueKeeper.DeleteWithdrawerMap(suite.ctx, withdraw, contract)
			},
			false,
		},
	}
	for _, tc := range testCases {
		tc.malleate()
		found := suite.app.RevenueKeeper.IsWithdrawerMapSet(suite.ctx, withdraw, contract)
		if tc.ok {
			suite.Require().True(found, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestIsRevenueRegistered() {
	feeSplit := types.NewRevenue(contract, deployer, withdraw)
	suite.app.RevenueKeeper.SetRevenue(suite.ctx, feeSplit)
	_, found := suite.app.RevenueKeeper.GetRevenue(suite.ctx, contract)
	suite.Require().True(found)

	testCases := []struct {
		name     string
		contract common.Address
		ok       bool
	}{
		{"registered fee split", contract, true},
		{"fee split not registered", common.Address{}, false},
		{"fee split not registered", tests.GenerateAddress(), false},
	}
	for _, tc := range testCases {
		found := suite.app.RevenueKeeper.IsRevenueRegistered(suite.ctx, tc.contract)
		if tc.ok {
			suite.Require().True(found, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}
