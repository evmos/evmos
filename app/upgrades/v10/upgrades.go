// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package v10

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v10
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakingKeeper stakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		setMinCommissionRate(ctx, stakingKeeper)

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// setMinCommissionRate sets the minimum commission rate for validators
// to 5%.
func setMinCommissionRate(ctx sdk.Context, sk stakingkeeper.Keeper) {
	stakingParams := stakingtypes.Params{
		UnbondingTime:     sk.UnbondingTime(ctx),
		MaxValidators:     sk.MaxValidators(ctx),
		MaxEntries:        sk.MaxEntries(ctx),
		HistoricalEntries: sk.HistoricalEntries(ctx),
		BondDenom:         sk.BondDenom(ctx),
		MinCommissionRate: sdk.NewDecWithPrec(5, 2), // 5%
	}

	sk.SetParams(ctx, stakingParams)
}
