// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v14

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	vestingkeeper "github.com/evmos/evmos/v14/x/vesting/keeper"
	vestingtypes "github.com/evmos/evmos/v14/x/vesting/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v14
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	vk vestingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// Set the default params for the vesting module
		if err := vk.SetParams(ctx, vestingtypes.DefaultParams()); err != nil {
			logger.Error("error while setting vesting parameters", "error", err)
		}

		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
