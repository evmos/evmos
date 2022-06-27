package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/tests"
	"github.com/evmos/evmos/v6/x/fees/types"
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
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)
				expRes = []types.Fee{
					types.NewFee(contract, deployer, withdraw),
				}
			},
		},
		{
			"one fee registered with no withdraw address",
			func() {
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, nil)
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
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)
				suite.app.FeesKeeper.SetFee(suite.ctx, contract2, deployer, nil)
				suite.app.FeesKeeper.SetFee(suite.ctx, contract3, deployer2, nil)
				expRes = []types.Fee{
					types.NewFee(contract, deployer, withdraw),
					types.NewFee(contract2, deployer, nil),
					types.NewFee(contract3, deployer2, nil),
				}
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
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)
				expRes = []types.Fee{
					types.NewFee(contract, deployer, withdraw),
				}
			},
		},
		{
			"one fee registered with no withdraw address",
			func() {
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, nil)
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
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)
				suite.app.FeesKeeper.SetFee(suite.ctx, contract2, deployer, nil)
				suite.app.FeesKeeper.SetFee(suite.ctx, contract3, deployer2, nil)
				expRes = []types.Fee{
					types.NewFee(contract, deployer, withdraw),
					types.NewFee(contract2, deployer, nil),
					types.NewFee(contract3, deployer2, nil),
				}
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
		malleate    func(common.Address, sdk.AccAddress, sdk.AccAddress)
		expDeployer bool
		expWithdraw bool
	}{
		{
			"fee with no withdraw address",
			contract,
			deployer,
			nil,
			func(contract common.Address, deployer sdk.AccAddress, withdraw sdk.AccAddress) {
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)
			},
			true,
			false,
		},
		{
			"fee with withdraw address same as deployer",
			contract,
			deployer,
			deployer,
			func(contract common.Address, deployer sdk.AccAddress, withdraw sdk.AccAddress) {
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)
			},
			true,
			false,
		},
		{
			"fee with withdraw address same as contract",
			contract,
			deployer,
			sdk.AccAddress(contract.Bytes()),
			func(contract common.Address, deployer sdk.AccAddress, withdraw sdk.AccAddress) {
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)
			},
			true,
			true,
		},
		{
			"fee with withdraw address different than deployer",
			contract,
			deployer,
			withdraw,
			func(contract common.Address, deployer sdk.AccAddress, withdraw sdk.AccAddress) {
				suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)
			},
			true,
			true,
		},
		{
			"no fee",
			common.Address{},
			nil,
			nil,
			func(contract common.Address, deployer sdk.AccAddress, withdraw sdk.AccAddress) {},
			false,
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			tc.malleate(tc.contract, tc.deployer, tc.withdraw)

			fee, found := suite.app.FeesKeeper.GetFee(suite.ctx, tc.contract)
			deployer, foundD := suite.app.FeesKeeper.GetDeployer(suite.ctx, tc.contract)
			withdraw, foundW := suite.app.FeesKeeper.GetWithdrawal(suite.ctx, tc.contract)

			if tc.expDeployer {
				suite.Require().True(found, tc.name)
				suite.Require().True(foundD, tc.name)
				suite.Require().Equal(tc.deployer, deployer, tc.name)
				suite.Require().Equal(tc.deployer.String(), fee.DeployerAddress, tc.name)
				suite.Require().Equal(tc.contract.Hex(), fee.ContractAddress, tc.name)

				if tc.expWithdraw {
					suite.Require().True(foundW, tc.name)
					suite.Require().Equal(tc.withdraw, withdraw, tc.name)
					suite.Require().Equal(tc.withdraw.String(), fee.WithdrawAddress, tc.name)
				} else {
					suite.Require().False(foundW, tc.name)
					suite.Require().Nil(withdraw, tc.name)
				}
			} else {
				suite.Require().False(found, tc.name)
				suite.Require().False(foundD, tc.name)
				suite.Require().Nil(deployer, tc.name)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDeleteFee() {
	// Register fee
	suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)

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
				suite.app.FeesKeeper.DeleteFee(suite.ctx, contract)
			},
			false,
		},
	}
	for _, tc := range testCases {
		tc.malleate()
		fee, found := suite.app.FeesKeeper.GetFee(suite.ctx, contract)
		d, foundD := suite.app.FeesKeeper.GetDeployer(suite.ctx, contract)
		w, foundW := suite.app.FeesKeeper.GetWithdrawal(suite.ctx, contract)
		if tc.ok {
			suite.Require().True(found, tc.name)
			suite.Require().True(foundD, tc.name)
			suite.Require().True(foundW, tc.name)
			suite.Require().Equal(initialFee, fee, tc.name)
			suite.Require().Equal(deployer, d, tc.name)
			suite.Require().Equal(withdraw, w, tc.name)
		} else {
			suite.Require().False(found, tc.name)
			suite.Require().False(foundD, tc.name)
			suite.Require().False(foundW, tc.name)
			suite.Require().Equal(types.Fee{}, fee, tc.name)
			suite.Require().Nil(d)
			suite.Require().Nil(w)
		}
	}
}

func (suite *KeeperTestSuite) TestIsFeeRegistered() {
	suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)
	_, found := suite.app.FeesKeeper.GetFee(suite.ctx, contract)
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
		found := suite.app.FeesKeeper.IsFeeRegistered(suite.ctx, tc.contract)
		if tc.ok {
			suite.Require().True(found, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestGetDeployerFees() {
	suite.app.FeesKeeper.SetDeployerFees(suite.ctx, deployer, contract)

	deployer2 := sdk.AccAddress(tests.GenerateAddress().Bytes())
	contract2 := tests.GenerateAddress()
	contract3 := tests.GenerateAddress()
	suite.app.FeesKeeper.SetDeployerFees(suite.ctx, deployer2, contract2)
	suite.app.FeesKeeper.SetDeployerFees(suite.ctx, deployer2, contract3)

	testCases := []struct {
		name        string
		deployer    sdk.AccAddress
		exitingFees []common.Address
	}{
		{"has registered contracts", deployer, []common.Address{contract}},
		{"has two registered contracts", deployer2, []common.Address{contract2, contract3}},
		{"has no registered contracts", sdk.AccAddress(tests.GenerateAddress().Bytes()), []common.Address{}},
	}
	for _, tc := range testCases {
		addresses := suite.app.FeesKeeper.GetDeployerFees(suite.ctx, tc.deployer)
		suite.Require().ElementsMatch(tc.exitingFees, addresses, tc.name)
	}
}

func (suite *KeeperTestSuite) TestDeleteDeployerFees() {
	contract2 := tests.GenerateAddress()
	setup := func() {
		suite.app.FeesKeeper.SetDeployerFees(suite.ctx, deployer, contract)
		suite.app.FeesKeeper.SetDeployerFees(suite.ctx, deployer, contract2)
	}

	testCases := []struct {
		name          string
		malleate      func()
		deletedFees   []common.Address
		remainingFees []common.Address
	}{
		{
			"existing fees, no delete",
			setup,
			[]common.Address{},
			[]common.Address{contract, contract2},
		},
		{
			"existing fees, delete one fee",
			func() {
				setup()
				suite.app.FeesKeeper.DeleteDeployerFees(suite.ctx, deployer, contract)
			},
			[]common.Address{contract},
			[]common.Address{contract2},
		},
		{
			"existing fees, delete all fees",
			func() {
				setup()
				suite.app.FeesKeeper.DeleteDeployerFees(suite.ctx, deployer, contract)
				suite.app.FeesKeeper.DeleteDeployerFees(suite.ctx, deployer, contract2)
			},
			[]common.Address{contract, contract2},
			[]common.Address{},
		},
		{
			"delete non existent fee",
			func() {
				setup()
				contract3 := tests.GenerateAddress()
				suite.app.FeesKeeper.DeleteDeployerFees(suite.ctx, deployer, contract3)
			},
			[]common.Address{},
			[]common.Address{contract, contract2},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			tc.malleate()
			for _, deletedFee := range tc.deletedFees {
				hasFee := suite.app.FeesKeeper.IsDeployerFeesRegistered(suite.ctx, deployer, deletedFee)
				suite.Require().False(hasFee, tc.name)
			}
			remainingFees := suite.app.FeesKeeper.GetDeployerFees(suite.ctx, deployer)
			suite.Require().ElementsMatch(tc.remainingFees, remainingFees, tc.name)
		})
	}
}

func (suite *KeeperTestSuite) TestIsDeployerFeesRegistered() {
	suite.app.FeesKeeper.SetDeployerFees(suite.ctx, deployer, contract)

	testCases := []struct {
		name     string
		deployer sdk.AccAddress
		contract common.Address
		ok       bool
	}{
		{"deployer has contract", deployer, contract, true},
		{"deployer does not have contract", sdk.AccAddress(tests.GenerateAddress().Bytes()), contract, false},
	}
	for _, tc := range testCases {
		found := suite.app.FeesKeeper.IsDeployerFeesRegistered(suite.ctx, tc.deployer, tc.contract)
		if tc.ok {
			suite.Require().True(found, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}
