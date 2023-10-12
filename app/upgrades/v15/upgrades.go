// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v15

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
<<<<<<< HEAD
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	v14 "github.com/evmos/evmos/v15/app/upgrades/v14"
	"github.com/evmos/evmos/v15/utils"
	evmkeeper "github.com/evmos/evmos/v15/x/evm/keeper"
=======
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
>>>>>>> a4569105 (imp: remove `crisis` module in v15 upgrade (#1847))
)

// CreateUpgradeHandler creates an SDK upgrade handler for v15.0.0
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
<<<<<<< HEAD
	bk bankkeeper.Keeper,
	ek *evmkeeper.Keeper,
	sk stakingkeeper.Keeper,
=======
>>>>>>> a4569105 (imp: remove `crisis` module in v15 upgrade (#1847))
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

<<<<<<< HEAD
		if utils.IsMainnet(ctx.ChainID()) {
			logger.Info("migrating strategic reserves")
			if err := v14.MigrateNativeMultisigs(
				ctx, bk, sk, v14.NewTeamStrategicReserveAcc, v14.OldStrategicReserves...,
			); err != nil {
				// NOTE: log error instead of aborting the upgrade
				logger.Error("error while migrating native multisigs", "error", err)
			}
		}

		// Add EIP contained in Shanghai hard fork to the extra EIPs
		// in the EVM parameters. This enables using the PUSH0 opcode and
		// thus supports Solidity v0.8.20.
		logger.Info("adding EIP 3855 to EVM parameters")
		err := EnableEIPs(ctx, ek, 3855)
		if err != nil {
			logger.Error("error while enabling EIPs", "error", err)
		}

		// we are deprecating the crisis module since it is not being used
		logger.Debug("deleting crisis module from version map...")
=======
		// we are depecrating crisis module since it is not being used
		logger.Debug("deleting feesplit module from version map...")
>>>>>>> a4569105 (imp: remove `crisis` module in v15 upgrade (#1847))
		delete(vm, "crisis")

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
<<<<<<< HEAD

// EnableEIPs enables the given EIPs in the EVM parameters.
func EnableEIPs(ctx sdk.Context, ek *evmkeeper.Keeper, eips ...int64) error {
	evmParams := ek.GetParams(ctx)
	evmParams.ExtraEIPs = append(evmParams.ExtraEIPs, eips...)

	return ek.SetParams(ctx, evmParams)
}
=======
>>>>>>> a4569105 (imp: remove `crisis` module in v15 upgrade (#1847))
