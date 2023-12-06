// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v16_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	v16 "github.com/evmos/evmos/v16/app/upgrades/v16"
	testnetwork "github.com/evmos/evmos/v16/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v16/utils"
)

func (its *IntegrationTestSuite) TestMigrateFeeCollector() {
	its.SetupTest()

	feeCollectorModuleAccount := its.network.App.AccountKeeper.GetModuleAccount(its.network.GetContext(), types.FeeCollectorName)
	modAcc, ok := feeCollectorModuleAccount.(*types.ModuleAccount)
	its.Require().True(ok)

	oldFeeCollector := types.NewModuleAccount(modAcc.BaseAccount, types.FeeCollectorName)

	its.Require().NotNil(oldFeeCollector)
	its.Require().Len(oldFeeCollector.GetPermissions(), 0)

	// Create a new FeeCollector module account with the same address and the new permissions.
	newFeeCollectorModuleAccount := types.NewModuleAccount(modAcc.BaseAccount, types.FeeCollectorName, types.Burner)
	its.network.App.AccountKeeper.SetModuleAccount(its.network.GetContext(), newFeeCollectorModuleAccount)

	newFeeCollector := its.network.App.AccountKeeper.GetModuleAccount(its.network.GetContext(), types.FeeCollectorName)
	its.Require().True(ok)
	its.Require().Equal(feeCollectorModuleAccount.GetAccountNumber(), newFeeCollector.GetAccountNumber())
	its.Require().Equal(feeCollectorModuleAccount.GetAddress(), newFeeCollector.GetAddress())
	its.Require().Equal(feeCollectorModuleAccount.GetName(), newFeeCollector.GetName())
	its.Require().Equal(feeCollectorModuleAccount.GetPermissions(), newFeeCollector.GetPermissions())
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
