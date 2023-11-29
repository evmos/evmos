// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v16

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	inflationkeeper "github.com/evmos/evmos/v15/x/inflation/v1/keeper"
)

// BurnUsageIncentivesPool burns the entirety of the usage incentives pool
func BurnUsageIncentivesPool(ctx sdk.Context, bk bankkeeper.Keeper) error {
	incentivesAddr := authtypes.NewModuleAddress("incentives")
	incentivesPoolBalance := bk.GetAllBalances(ctx, incentivesAddr)
	if !incentivesPoolBalance.IsAllPositive() {
		return nil
	}

	// transfer funds to the fee collector account and then burn it
	if err := bk.SendCoinsFromAccountToModule(ctx, incentivesAddr, authtypes.FeeCollectorName, incentivesPoolBalance); err != nil {
		return err
	}

	return bk.BurnCoins(ctx, authtypes.FeeCollectorName, incentivesPoolBalance)
}

// UpdateInflationParams updates the inflation params to adjust the inflation distribution while removing
// the usage incentive allocation portion of it.
func UpdateInflationParams(ctx sdk.Context, ik inflationkeeper.Keeper) error {
	params := ik.GetParams(ctx)
	params.InflationDistribution.CommunityPool = sdkmath.LegacyOneDec().Sub(params.InflationDistribution.StakingRewards)
	params.InflationDistribution.UsageIncentives = sdkmath.LegacyZeroDec() //nolint:staticcheck,nolintlint

	if err := params.Validate(); err != nil {
		return err
	}
	return ik.SetParams(ctx, params)
}
