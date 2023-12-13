// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v16_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v16/app/upgrades/v16/incentives"
	"github.com/evmos/evmos/v16/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v16/testutil"

	v16 "github.com/evmos/evmos/v16/app/upgrades/v16"
	testnetwork "github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v16/utils"
)

func (its *IntegrationTestSuite) TestProposalDeletion() {
	its.SetupTest()

	proposal := &incentives.RegisterIncentiveProposal{
		Title:       "Test",
		Description: "Test Incentive Proposal",
		Contract:    "",
		Allocations: nil,
		Epochs:      100,
	}
	privKey, _ := ethsecp256k1.GenerateKey()
	addrBz := privKey.PubKey().Address().Bytes()
	accAddr := sdk.AccAddress(addrBz)
	coins := sdk.NewCoins(sdk.NewCoin(its.network.GetDenom(), math.NewInt(100000000)))
	err := testutil.FundAccount(its.network.GetContext(), its.network.App.BankKeeper, accAddr, coins)
	its.Require().NoError(err)

	_, err = testutil.SubmitProposal(its.network.GetContext(), its.network.App, privKey, proposal, 0)
	its.Require().NoError(err)

}

func (its *IntegrationTestSuite) TestBurnUsageIncentivesPool() {
	its.SetupTest()
	// check initial balance of incentives mod account
	expIntialBalance := sdk.NewCoin(utils.BaseDenom, testnetwork.PrefundedAccountInitialBalance)
	res, err := its.handler.GetBalance(its.incentivesAcc, utils.BaseDenom)
	its.Require().NoError(err)
	its.Require().NotNil(res.Balance)
	its.Require().Equal(expIntialBalance, *res.Balance)

	err = v16.BurnUsageIncentivesPool(its.network.GetContext(), its.network.App.BankKeeper)
	its.Require().NoError(err)

	// Check incentives pool final balance - should be zero
	expFinalBalance := sdk.NewCoin(utils.BaseDenom, math.ZeroInt())
	res, err = its.handler.GetBalance(its.incentivesAcc, utils.BaseDenom)
	its.Require().NoError(err)
	its.Require().NotNil(res.Balance)
	its.Require().Equal(expFinalBalance, *res.Balance)
}

func (its *IntegrationTestSuite) TestUpdateInflationParams() {
	its.SetupTest()
	// check initial inflation params has incentive allocation > 0
	initialParams := its.network.App.InflationKeeper.GetParams(its.network.GetContext())
	its.Require().Equal(initialParams.InflationDistribution.UsageIncentives, initialInflDistr.UsageIncentives) //nolint:staticcheck

	err := v16.UpdateInflationParams(its.network.GetContext(), its.network.App.InflationKeeper)
	its.Require().NoError(err)

	// Check incentives allocation is zero
	finalParams := its.network.App.InflationKeeper.GetParams(its.network.GetContext())
	its.Require().Equal(math.LegacyZeroDec(), finalParams.InflationDistribution.UsageIncentives) //nolint:staticcheck
}
