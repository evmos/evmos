// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v16_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/evmos/evmos/v16/app/upgrades/v16/incentives"
	"github.com/evmos/evmos/v16/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v16/testutil"
	utiltx "github.com/evmos/evmos/v16/testutil/tx"
	"time"

	v16 "github.com/evmos/evmos/v16/app/upgrades/v16"
	testnetwork "github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v16/utils"
)

func (its *IntegrationTestSuite) TestProposalDeletion() {
	its.SetupTest()
	incentives.RegisterInterfaces(its.network.App.InterfaceRegistry())

	proposal := &incentives.RegisterIncentiveProposal{
		Title:       "Test",
		Description: "Test Incentive Proposal",
		Contract:    utiltx.GenerateAddress().String(),
		Allocations: sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
		Epochs:      100,
	}
	privKey, _ := ethsecp256k1.GenerateKey()
	addrBz := privKey.PubKey().Address().Bytes()
	accAddr := sdk.AccAddress(addrBz)
	coins := sdk.NewCoins(sdk.NewCoin(its.network.GetDenom(), math.NewInt(5e18)))
	err := testutil.FundAccount(its.network.GetContext(), its.network.App.BankKeeper, accAddr, coins)
	its.Require().NoError(err)

	content, err := govtypesv1.NewLegacyContent(
		proposal,
		sdk.MustBech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix(), accAddr),
	)

	proposalMsgs := []sdk.Msg{content}
	newProposal, err := govtypesv1.NewProposal(proposalMsgs, 1, time.Now(), time.Now().Add(time.Hour*5), "", "Test", "Test", accAddr)
	its.Require().NoError(err)
	its.network.App.GovKeeper.SetProposal(its.network.GetContext(), newProposal)

	allProposalsBefore := its.network.App.GovKeeper.GetProposals(its.network.GetContext())
	its.Require().Len(allProposalsBefore, 1)

	logger := its.network.GetContext().Logger()
	v16.DeleteRegisterIncentivesProposals(its.network.GetContext(), its.network.App.GovKeeper, logger)

	allProposalsAfter := its.network.App.GovKeeper.GetProposals(its.network.GetContext())
	its.Require().Len(allProposalsAfter, 0)

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
