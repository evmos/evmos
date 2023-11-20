// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v10

import (
	"context"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v10
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakingKeeper stakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		logger := ctx.Logger().With("upgrade", UpgradeName)

		if err := setMinCommissionRate(ctx, stakingKeeper); err != nil {
			return nil, err
		}

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// setMinCommissionRate sets the minimum commission rate for validators
// to 5%.
func setMinCommissionRate(ctx sdk.Context, sk stakingkeeper.Keeper) error {
	unbondingTime, err := sk.UnbondingTime(ctx)
	if err != nil {
		return err
	}

	maxValidators, err := sk.MaxValidators(ctx)
	if err != nil {
		return err
	}

	maxEntries, err := sk.MaxEntries(ctx)
	if err != nil {
		return err
	}

	historicalEntries, err := sk.HistoricalEntries(ctx)
	if err != nil {
		return err
	}

	bondDenom, err := sk.BondDenom(ctx)
	if err != nil {
		return err
	}

	stakingParams := stakingtypes.Params{
		UnbondingTime:     unbondingTime,
		MaxValidators:     maxValidators,
		MaxEntries:        maxEntries,
		HistoricalEntries: historicalEntries,
		BondDenom:         bondDenom,
		MinCommissionRate: math.LegacyNewDecWithPrec(5, 2), // 5%
	}

	return sk.SetParams(ctx, stakingParams)
}
