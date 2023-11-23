// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v16

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	incentiveskeeper "github.com/evmos/evmos/v15/x/incentives/keeper"
	incentivestypes "github.com/evmos/evmos/v15/x/incentives/types"
	inflationkeeper "github.com/evmos/evmos/v15/x/inflation/v1/keeper"
)

// BurnUsageIncentivesPool burns the entirety of the usage incentives pool
func BurnUsageIncentivesPool(ctx sdk.Context, bk bankkeeper.Keeper) error {
	incentivesPoolBalance := bk.GetAllBalances(ctx, authtypes.NewModuleAddress(incentivestypes.ModuleName))
	if !incentivesPoolBalance.IsAllPositive() {
		return nil
	}

	return bk.BurnCoins(ctx, authtypes.FeeCollectorName, incentivesPoolBalance)
}

// DisableUsageIncentives disables the usage incentives
func DisableUsageIncentives(ctx sdk.Context, incentivesKeeper incentiveskeeper.Keeper) error {
	params := incentivesKeeper.GetParams(ctx)
	params.EnableIncentives = false
	return incentivesKeeper.SetParams(ctx, params)
}

// UpdateInflationParams updates the inflation params to and sets the usage incentive allocation
// to zero.
func UpdateInflationParams(ctx sdk.Context, ik inflationkeeper.Keeper) error {
	params := ik.GetParams(ctx)
	params.InflationDistribution.CommunityPool = sdkmath.LegacyOneDec().Sub(params.InflationDistribution.StakingRewards)
	params.InflationDistribution.UsageIncentives = sdkmath.LegacyZeroDec() // set the usage incentive to zero

	if err := params.Validate(); err != nil {
		return err
	}
	return ik.SetParams(ctx, params)
}
