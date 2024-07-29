// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v16_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	v16 "github.com/evmos/evmos/v19/app/upgrades/v16"
	"github.com/evmos/evmos/v19/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v19/testutil"
	testnetwork "github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	utiltx "github.com/evmos/evmos/v19/testutil/tx"
	"github.com/evmos/evmos/v19/utils"
	incentives "github.com/evmos/evmos/v19/x/incentives/types"
)

func (its *IntegrationTestSuite) TestFeeCollectorMigration() {
	its.SetupTest()
	context := its.network.GetContext()

	// get current fee collector
	feeCollectorModuleAccount := its.network.App.AccountKeeper.GetModuleAccount(context, authtypes.FeeCollectorName)

	modAcc, ok := feeCollectorModuleAccount.(*authtypes.ModuleAccount)
	its.Require().Equal(true, ok)

	// save fee collector without burner auth
	feeCollectorModuleAccountNoBurner := authtypes.NewModuleAccount(modAcc.BaseAccount, authtypes.FeeCollectorName)
	its.network.App.AccountKeeper.SetModuleAccount(context, feeCollectorModuleAccountNoBurner)

	// check fee collector is without burner auth
	feeCollectorNoBurner := its.network.App.AccountKeeper.GetModuleAccount(context, authtypes.FeeCollectorName)
	hasBurnerPermission := feeCollectorNoBurner.HasPermission(authtypes.Burner)
	its.Require().True(!hasBurnerPermission)

	err := v16.MigrateFeeCollector(its.network.App.AccountKeeper, context)
	its.Require().NoError(err)

	// check fee collector has burner permission
	feeCollectorAfterMigration := its.network.App.AccountKeeper.GetModuleAccount(context, authtypes.FeeCollectorName)
	hasBurnerPermission = feeCollectorAfterMigration.HasPermission(authtypes.Burner)
	its.Require().True(hasBurnerPermission)
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

func (its *IntegrationTestSuite) TestDeleteIncentivesProposals() {
	its.SetupTest()

	// Create 3 proposals. 2 will be deleted because correspond to the incentives module
	expInitialProps := 3
	expFinalProps := 1
	prop1 := &incentives.RegisterIncentiveProposal{
		Title:       "Test",
		Description: "Test Register Incentive Proposal",
		Contract:    utiltx.GenerateAddress().String(),
		Allocations: sdk.DecCoins{sdk.NewDecCoinFromDec("aevmos", sdk.NewDecWithPrec(5, 2))},
		Epochs:      100,
	}

	prop2 := &upgrade.SoftwareUpgradeProposal{ //nolint:staticcheck
		Title:       "Test",
		Description: "Test Software Upgrade Proposal",
		Plan:        upgrade.Plan{},
	}

	prop3 := &incentives.CancelIncentiveProposal{
		Title:       "Test",
		Description: "Test Cancel Incentive Proposal",
		Contract:    utiltx.GenerateAddress().String(),
	}

	privKey, _ := ethsecp256k1.GenerateKey()
	addrBz := privKey.PubKey().Address().Bytes()
	accAddr := sdk.AccAddress(addrBz)
	coins := sdk.NewCoins(sdk.NewCoin(its.network.GetDenom(), math.NewInt(5e18)))
	err := testutil.FundAccount(its.network.GetContext(), its.network.App.BankKeeper, accAddr, coins)
	its.Require().NoError(err)

	for _, prop := range []govtypesv1beta.Content{prop1, prop2, prop3} {
		its.createProposal(prop, accAddr)
	}

	// check the creation of the 3 proposals was successful
	allProposalsBefore := its.network.App.GovKeeper.GetProposals(its.network.GetContext())
	its.Require().Len(allProposalsBefore, expInitialProps)

	// Delete the corresponding proposals
	logger := its.network.GetContext().Logger()
	v16.DeleteIncentivesProposals(its.network.GetContext(), its.network.App.GovKeeper, logger)

	allProposalsAfter := its.network.App.GovKeeper.GetProposals(its.network.GetContext())
	its.Require().Len(allProposalsAfter, expFinalProps)
}

func (its *IntegrationTestSuite) createProposal(content govtypesv1beta.Content, acc sdk.AccAddress) {
	allProposalsBefore := its.network.App.GovKeeper.GetProposals(its.network.GetContext())
	propID := len(allProposalsBefore) + 1

	legacyContent, err := govtypesv1.NewLegacyContent(
		content,
		sdk.MustBech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix(), acc),
	)
	its.Require().NoError(err)

	proposalMsgs := []sdk.Msg{legacyContent}
	newProposal, err := govtypesv1.NewProposal(proposalMsgs, uint64(propID), time.Now(), time.Now().Add(time.Hour*5), "", "Test", "Test", acc)
	its.Require().NoError(err)
	its.network.App.GovKeeper.SetProposal(its.network.GetContext(), newProposal)
}
