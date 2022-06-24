package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v5/x/fees/types"
	"github.com/tharsis/ethermint/tests"
)

func (suite *KeeperTestSuite) TestGetFees() {
	var expRes []types.Fee

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"no fees registered",
			func() { expRes = []types.Fee{} },
		},
		{
			"one fee registered with withdraw address",
			func() {
				fee := types.NewFee(contract, deployer, withdraw)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee)
				expRes = []types.Fee{fee}
			},
		},
		{
			"one fee registered with no withdraw address",
			func() {
				fee := types.NewFee(contract, deployer, nil)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee)
				expRes = []types.Fee{fee}
			},
		},
		{
			"multiple fees registered",
			func() {
				deployer2 := sdk.AccAddress(tests.GenerateAddress().Bytes())
				contract2 := tests.GenerateAddress()
				contract3 := tests.GenerateAddress()
				fee := types.NewFee(contract, deployer, withdraw)
				fee2 := types.NewFee(contract2, deployer, nil)
				fee3 := types.NewFee(contract3, deployer2, nil)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee2)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee3)
				expRes = []types.Fee{fee, fee2, fee3}
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			tc.malleate()

			res := suite.app.FeesKeeper.GetFees(suite.ctx)
			suite.Require().ElementsMatch(expRes, res, tc.name)
		})
	}
}

func (suite *KeeperTestSuite) TestIterateFees() {
	var expRes []types.Fee

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"no fees registered",
			func() { expRes = []types.Fee{} },
		},
		{
			"one fee registered with withdraw address",
			func() {
				fee := types.NewFee(contract, deployer, withdraw)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee)
				expRes = []types.Fee{
					types.NewFee(contract, deployer, withdraw),
				}
			},
		},
		{
			"one fee registered with no withdraw address",
			func() {
				fee := types.NewFee(contract, deployer, nil)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee)
				expRes = []types.Fee{
					types.NewFee(contract, deployer, nil),
				}
			},
		},
		{
			"multiple fees registered",
			func() {
				deployer2 := sdk.AccAddress(tests.GenerateAddress().Bytes())
				contract2 := tests.GenerateAddress()
				contract3 := tests.GenerateAddress()
				fee := types.NewFee(contract, deployer, withdraw)
				fee2 := types.NewFee(contract2, deployer, nil)
				fee3 := types.NewFee(contract3, deployer2, nil)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee2)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee3)
				expRes = []types.Fee{fee, fee2, fee3}
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			tc.malleate()

			suite.app.FeesKeeper.IterateFees(suite.ctx, func(fee types.Fee) (stop bool) {
				suite.Require().Contains(expRes, fee, tc.name)
				return false
			})
		})
	}
}

func (suite *KeeperTestSuite) TestGetFee() {
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
				fee := types.NewFee(tc.contract, tc.deployer, tc.withdraw)
				suite.app.FeesKeeper.SetFee(suite.ctx, fee)
				suite.app.FeesKeeper.SetDeployerMap(suite.ctx, tc.deployer, tc.contract)
			}

			if tc.expWithdraw {
				suite.app.FeesKeeper.SetWithdrawMap(suite.ctx, tc.withdraw, tc.contract)
			}

			fee, found := suite.app.FeesKeeper.GetFee(suite.ctx, tc.contract)
			foundD := suite.app.FeesKeeper.IsDeployerMapSet(suite.ctx, tc.deployer, tc.contract)
			foundW := suite.app.FeesKeeper.IsWithdrawMapSet(suite.ctx, tc.withdraw, tc.contract)

			if tc.found {
				suite.Require().True(found, tc.name)
				suite.Require().Equal(tc.deployer.String(), fee.DeployerAddress, tc.name)
				suite.Require().Equal(tc.contract.Hex(), fee.ContractAddress, tc.name)

				suite.Require().True(foundD, tc.name)

				if tc.expWithdraw {
					suite.Require().Equal(tc.withdraw.String(), fee.WithdrawAddress, tc.name)
					suite.Require().True(foundW, tc.name)
				} else {
					suite.Require().Equal("", fee.WithdrawAddress, tc.name)
					suite.Require().False(foundW, tc.name)
				}
			} else {
				suite.Require().False(found, tc.name)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDeleteFee() {
	// Register fee
	fee := types.NewFee(contract, deployer, withdraw)
	suite.app.FeesKeeper.SetFee(suite.ctx, fee)

	initialFee, found := suite.app.FeesKeeper.GetFee(suite.ctx, contract)
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
				suite.app.FeesKeeper.DeleteFee(suite.ctx, fee)
			},
			false,
		},
	}
	for _, tc := range testCases {
		tc.malleate()
		fee, found := suite.app.FeesKeeper.GetFee(suite.ctx, contract)
		if tc.ok {
			suite.Require().True(found, tc.name)
			suite.Require().Equal(initialFee, fee, tc.name)
		} else {
			suite.Require().False(found, tc.name)
			suite.Require().Equal(types.Fee{}, fee, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestDeleteDeployerMap() {
	suite.app.FeesKeeper.SetDeployerMap(suite.ctx, deployer, contract)
	found := suite.app.FeesKeeper.IsDeployerMapSet(suite.ctx, deployer, contract)
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
				suite.app.FeesKeeper.DeleteDeployerMap(suite.ctx, deployer, contract)
			},
			false,
		},
	}
	for _, tc := range testCases {
		tc.malleate()
		found := suite.app.FeesKeeper.IsDeployerMapSet(suite.ctx, deployer, contract)
		if tc.ok {
			suite.Require().True(found, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestDeleteWithdrawMap() {
	suite.app.FeesKeeper.SetWithdrawMap(suite.ctx, withdraw, contract)
	found := suite.app.FeesKeeper.IsWithdrawMapSet(suite.ctx, withdraw, contract)
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
				suite.app.FeesKeeper.DeleteWithdrawMap(suite.ctx, withdraw, contract)
			},
			false,
		},
	}
	for _, tc := range testCases {
		tc.malleate()
		found := suite.app.FeesKeeper.IsWithdrawMapSet(suite.ctx, withdraw, contract)
		if tc.ok {
			suite.Require().True(found, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}
