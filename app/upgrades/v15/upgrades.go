// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v15

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	v14 "github.com/evmos/evmos/v14/app/upgrades/v14"
	"github.com/evmos/evmos/v14/utils"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v15.0.0
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bk bankkeeper.Keeper,
	sk stakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

<<<<<<< HEAD:app/upgrades/v15/upgrades.go
		// we are depecrating crisis module since it is not being used
		logger.Debug("deleting feesplit module from version map...")
		delete(vm, "crisis")
=======
		if utils.IsMainnet(ctx.ChainID()) {
			logger.Info("migrating strategic reserves")
			if err := v14.MigrateNativeMultisigs(
				ctx, bk, sk, v14.NewTeamStrategicReserveAcc, v14.OldStrategicReserves...,
			); err != nil {
				// NOTE: log error instead of aborting the upgrade
				logger.Error("error while migrating native multisigs", "error", err)
			}
		}
>>>>>>> fc2c9a2a (chore(upgrade): add migration to upgrade logic (#1845)):app/upgrades/v14_2/upgrades.go

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
