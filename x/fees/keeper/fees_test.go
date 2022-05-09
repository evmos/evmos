package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/v4/x/fees/types"
)

var (
	contract = tests.GenerateAddress()
	deployer = sdk.AccAddress(tests.GenerateAddress().Bytes())
	withdraw = sdk.AccAddress(tests.GenerateAddress().Bytes())
)

func (suite *KeeperTestSuite) TestGetFeeInfo() {
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
			"fee with widthdraw address same as deployer",
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
			"fee with widthdraw address same as contract",
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
			"fee with widthdraw address different than deployer",
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
		suite.SetupTest() // reset
		tc.malleate(tc.contract, tc.deployer, tc.withdraw)

		fee, found := suite.app.FeesKeeper.GetFeeInfo(suite.ctx, tc.contract)
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
	}
}

func (suite *KeeperTestSuite) TestDeleteFee() {
	// Register fee
	suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)

	initialFee, found := suite.app.FeesKeeper.GetFeeInfo(suite.ctx, contract)
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
		fee, found := suite.app.FeesKeeper.GetFeeInfo(suite.ctx, contract)
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
			suite.Require().Equal(types.DevFeeInfo{}, fee, tc.name)
			suite.Require().Nil(d)
			suite.Require().Nil(w)
		}
	}
}

func (suite *KeeperTestSuite) TestIsFeeRegistered() {
	suite.app.FeesKeeper.SetFee(suite.ctx, contract, deployer, withdraw)
	_, found := suite.app.FeesKeeper.GetFeeInfo(suite.ctx, contract)
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

func (suite *KeeperTestSuite) TestGetFeesInverse() {

}

func (suite *KeeperTestSuite) TestDeleteFeeInverse() {
	// Register inverse fee for deployer
	suite.app.FeesKeeper.SetFeeInverse(suite.ctx, deployer, contract)

	testCases := []struct {
		name     string
		malleate func()
		ok       bool
	}{
		{"existing inverse fee mapping", func() {}, true},
		{
			"no inverse fee mapping",
			func() {
				suite.app.FeesKeeper.DeleteFeeInverse(suite.ctx, deployer)
			},
			false,
		},
	}
	for _, tc := range testCases {
		tc.malleate()
		hasFees := suite.app.FeesKeeper.HasFeeInverse(suite.ctx, deployer)
		addresses := suite.app.FeesKeeper.GetFeesInverse(suite.ctx, deployer)
		if tc.ok {
			suite.Require().True(hasFees, tc.name)
			suite.Require().Equal(len(addresses), 1, tc.name)
		} else {
			suite.Require().False(hasFees, tc.name)
			suite.Require().Equal(len(addresses), 0, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestHasFeeInverse() {
	suite.app.FeesKeeper.SetFeeInverse(suite.ctx, deployer, contract)

	testCases := []struct {
		name     string
		deployer sdk.AccAddress
		ok       bool
	}{
		{"deployer has fees", deployer, true},
		{"deployer does not have fees", sdk.AccAddress(tests.GenerateAddress().Bytes()), false},
	}
	for _, tc := range testCases {
		found := suite.app.FeesKeeper.HasFeeInverse(suite.ctx, tc.deployer)
		if tc.ok {
			suite.Require().True(found, tc.name)
		} else {
			suite.Require().False(found, tc.name)
		}
	}
}
