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

// func (suite *KeeperTestSuite) TestGetDeployerFees() {
// 	suite.app.FeesKeeper.SetDeployerFees(suite.ctx, deployer, contract)

// 	deployer2 := sdk.AccAddress(tests.GenerateAddress().Bytes())
// 	contract2 := tests.GenerateAddress()
// 	contract3 := tests.GenerateAddress()
// 	suite.app.FeesKeeper.SetDeployerFees(suite.ctx, deployer2, contract2)
// 	suite.app.FeesKeeper.SetDeployerFees(suite.ctx, deployer2, contract3)

// 	testCases := []struct {
// 		name        string
// 		deployer    sdk.AccAddress
// 		exitingFees []common.Address
// 	}{
// 		{"has registered contracts", deployer, []common.Address{contract}},
// 		{"has two registered contracts", deployer2, []common.Address{contract2, contract3}},
// 		{"has no registered contracts", sdk.AccAddress(tests.GenerateAddress().Bytes()), []common.Address{}},
// 	}
// 	for _, tc := range testCases {
// 		addresses := suite.app.FeesKeeper.GetDeployerFees(suite.ctx, tc.deployer)
// 		suite.Require().ElementsMatch(tc.exitingFees, addresses, tc.name)
// 	}
// }

// func (suite *KeeperTestSuite) TestDeleteDeployerFees() {
// 	contract2 := tests.GenerateAddress()
// 	setup := func() {
// 		suite.app.FeesKeeper.SetDeployerFees(suite.ctx, deployer, contract)
// 		suite.app.FeesKeeper.SetDeployerFees(suite.ctx, deployer, contract2)
// 	}

// 	testCases := []struct {
// 		name          string
// 		malleate      func()
// 		deletedFees   []common.Address
// 		remainingFees []common.Address
// 	}{
// 		{
// 			"existing fees, no delete",
// 			setup,
// 			[]common.Address{},
// 			[]common.Address{contract, contract2},
// 		},
// 		{
// 			"existing fees, delete one fee",
// 			func() {
// 				setup()
// 				suite.app.FeesKeeper.DeleteDeployerFees(suite.ctx, deployer, contract)
// 			},
// 			[]common.Address{contract},
// 			[]common.Address{contract2},
// 		},
// 		{
// 			"existing fees, delete all fees",
// 			func() {
// 				setup()
// 				suite.app.FeesKeeper.DeleteDeployerFees(suite.ctx, deployer, contract)
// 				suite.app.FeesKeeper.DeleteDeployerFees(suite.ctx, deployer, contract2)
// 			},
// 			[]common.Address{contract, contract2},
// 			[]common.Address{},
// 		},
// 		{
// 			"delete non existent fee",
// 			func() {
// 				setup()
// 				contract3 := tests.GenerateAddress()
// 				suite.app.FeesKeeper.DeleteDeployerFees(suite.ctx, deployer, contract3)
// 			},
// 			[]common.Address{},
// 			[]common.Address{contract, contract2},
// 		},
// 	}
// 	for _, tc := range testCases {
// 		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
// 			suite.SetupTest() // reset
// 			tc.malleate()
// 			for _, deletedFee := range tc.deletedFees {
// 				hasFee := suite.app.FeesKeeper.IsDeployerFeesRegistered(suite.ctx, deployer, deletedFee)
// 				suite.Require().False(hasFee, tc.name)
// 			}
// 			remainingFees := suite.app.FeesKeeper.GetDeployerFees(suite.ctx, deployer)
// 			suite.Require().ElementsMatch(tc.remainingFees, remainingFees, tc.name)
// 		})
// 	}
// }

// func (suite *KeeperTestSuite) TestIsDeployerFeesRegistered() {
// 	suite.app.FeesKeeper.SetDeployerFees(suite.ctx, deployer, contract)

// 	testCases := []struct {
// 		name     string
// 		deployer sdk.AccAddress
// 		contract common.Address
// 		ok       bool
// 	}{
// 		{"deployer has contract", deployer, contract, true},
// 		{"deployer does not have contract", sdk.AccAddress(tests.GenerateAddress().Bytes()), contract, false},
// 	}
// 	for _, tc := range testCases {
// 		found := suite.app.FeesKeeper.IsDeployerFeesRegistered(suite.ctx, tc.deployer, tc.contract)
// 		if tc.ok {
// 			suite.Require().True(found, tc.name)
// 		} else {
// 			suite.Require().False(found, tc.name)
// 		}
// 	}
// }
