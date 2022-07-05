package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/tests"
	"github.com/evmos/evmos/v6/x/feesplit/types"
)

func (suite *KeeperTestSuite) TestGetFees() {
	var expRes []types.FeeSplit

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"no fees registered",
			func() { expRes = []types.FeeSplit{} },
		},
		{
			"one fee registered with withdraw address",
			func() {
				fee := types.NewFeeSplit(contract, deployer, withdraw)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, fee)
				expRes = []types.FeeSplit{fee}
			},
		},
		{
			"one fee registered with no withdraw address",
			func() {
				fee := types.NewFeeSplit(contract, deployer, nil)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, fee)
				expRes = []types.FeeSplit{fee}
			},
		},
		{
			"multiple fees registered",
			func() {
				deployer2 := sdk.AccAddress(tests.GenerateAddress().Bytes())
				contract2 := tests.GenerateAddress()
				contract3 := tests.GenerateAddress()
				fee := types.NewFeeSplit(contract, deployer, withdraw)
				fee2 := types.NewFeeSplit(contract2, deployer, nil)
				fee3 := types.NewFeeSplit(contract3, deployer2, nil)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, fee)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, fee2)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, fee3)
				expRes = []types.FeeSplit{fee, fee2, fee3}
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			tc.malleate()

			res := suite.app.FeesplitKeeper.GetFeeSplits(suite.ctx)
			suite.Require().ElementsMatch(expRes, res, tc.name)
		})
	}
}

func (suite *KeeperTestSuite) TestIterateFees() {
	var expRes []types.FeeSplit

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"no fees registered",
			func() { expRes = []types.FeeSplit{} },
		},
		{
			"one fee registered with withdraw address",
			func() {
				fee := types.NewFeeSplit(contract, deployer, withdraw)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, fee)
				expRes = []types.FeeSplit{
					types.NewFeeSplit(contract, deployer, withdraw),
				}
			},
		},
		{
			"one fee registered with no withdraw address",
			func() {
				fee := types.NewFeeSplit(contract, deployer, nil)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, fee)
				expRes = []types.FeeSplit{
					types.NewFeeSplit(contract, deployer, nil),
				}
			},
		},
		{
			"multiple fees registered",
			func() {
				deployer2 := sdk.AccAddress(tests.GenerateAddress().Bytes())
				contract2 := tests.GenerateAddress()
				contract3 := tests.GenerateAddress()
				fee := types.NewFeeSplit(contract, deployer, withdraw)
				fee2 := types.NewFeeSplit(contract2, deployer, nil)
				fee3 := types.NewFeeSplit(contract3, deployer2, nil)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, fee)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, fee2)
				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, fee3)
				expRes = []types.FeeSplit{fee, fee2, fee3}
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			tc.malleate()

			suite.app.FeesplitKeeper.IterateFeeSplits(suite.ctx, func(fee types.FeeSplit) (stop bool) {
				suite.Require().Contains(expRes, fee, tc.name)
				return false
			})
		})
	}
}

func (suite *KeeperTestSuite) TestGetFeeSplit() {
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
				fee := types.NewFeeSplit(tc.contract, tc.deployer, tc.withdraw)
				if tc.deployer.Equals(tc.withdraw) {
					fee.WithdrawerAddress = ""
				}

				suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, fee)
				suite.app.FeesplitKeeper.SetDeployerMap(suite.ctx, tc.deployer, tc.contract)
			}

			if tc.expWithdraw {
				suite.app.FeesplitKeeper.SetWithdrawerMap(suite.ctx, tc.withdraw, tc.contract)
			}

			fee, found := suite.app.FeesplitKeeper.GetFeeSplit(suite.ctx, tc.contract)
			foundD := suite.app.FeesplitKeeper.IsDeployerMapSet(suite.ctx, tc.deployer, tc.contract)
			foundW := suite.app.FeesplitKeeper.IsWithdrawerMapSet(suite.ctx, tc.withdraw, tc.contract)

			if tc.found {
				suite.Require().True(found, tc.name)
				suite.Require().Equal(tc.deployer.String(), fee.DeployerAddress, tc.name)
				suite.Require().Equal(tc.contract.Hex(), fee.ContractAddress, tc.name)

				suite.Require().True(foundD, tc.name)

				if tc.expWithdraw {
					suite.Require().Equal(tc.withdraw.String(), fee.WithdrawerAddress, tc.name)
					suite.Require().True(foundW, tc.name)
				} else {
					suite.Require().Equal("", fee.WithdrawerAddress, tc.name)
					suite.Require().False(foundW, tc.name)
				}
			} else {
				suite.Require().False(found, tc.name)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDeleteFeeSplit() {
	fee := types.NewFeeSplit(contract, deployer, withdraw)
	suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, fee)

	initialFee, found := suite.app.FeesplitKeeper.GetFeeSplit(suite.ctx, contract)
	suite.Require().True(found)

	testCases := []struct {
		name     string
		malleate func()
		ok       bool
	}{
		{"existing fee", func() {}, true},
		{
			"deleted fee",
			func() {
				suite.app.FeesplitKeeper.DeleteFeeSplit(suite.ctx, fee)
			},
			false,
		},
	}
	for _, tc := range testCases {
		tc.malleate()
		fee, found := suite.app.FeesplitKeeper.GetFeeSplit(suite.ctx, contract)
		if tc.ok {
			suite.Require().True(found, tc.name)
			suite.Require().Equal(initialFee, fee, tc.name)
		} else {
			suite.Require().False(found, tc.name)
			suite.Require().Equal(types.FeeSplit{}, fee, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestDeleteDeployerMap() {
	suite.app.FeesplitKeeper.SetDeployerMap(suite.ctx, deployer, contract)
	found := suite.app.FeesplitKeeper.IsDeployerMapSet(suite.ctx, deployer, contract)
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
				suite.app.FeesplitKeeper.DeleteDeployerMap(suite.ctx, deployer, contract)
			},
			false,
		},
	}
	for _, tc := range testCases {
		tc.malleate()
		found := suite.app.FeesplitKeeper.IsDeployerMapSet(suite.ctx, deployer, contract)
		if tc.ok {
			suite.Require().True(found, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestDeleteWithdrawMap() {
	suite.app.FeesplitKeeper.SetWithdrawerMap(suite.ctx, withdraw, contract)
	found := suite.app.FeesplitKeeper.IsWithdrawerMapSet(suite.ctx, withdraw, contract)
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
				suite.app.FeesplitKeeper.DeleteWithdrawerMap(suite.ctx, withdraw, contract)
			},
			false,
		},
	}
	for _, tc := range testCases {
		tc.malleate()
		found := suite.app.FeesplitKeeper.IsWithdrawerMapSet(suite.ctx, withdraw, contract)
		if tc.ok {
			suite.Require().True(found, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestIsFeeSplitRegistered() {
	fee := types.NewFeeSplit(contract, deployer, withdraw)
	suite.app.FeesplitKeeper.SetFeeSplit(suite.ctx, fee)
	_, found := suite.app.FeesplitKeeper.GetFeeSplit(suite.ctx, contract)
	suite.Require().True(found)

	testCases := []struct {
		name     string
		contract common.Address
		ok       bool
	}{
		{"registered fee", contract, true},
		{"fee not registered", common.Address{}, false},
		{"fee not registered", tests.GenerateAddress(), false},
	}
	for _, tc := range testCases {
		found := suite.app.FeesplitKeeper.IsFeeSplitRegistered(suite.ctx, tc.contract)
		if tc.ok {
			suite.Require().True(found, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}
