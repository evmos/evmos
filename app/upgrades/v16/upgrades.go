// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v16

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	erc20keeper "github.com/evmos/evmos/v14/x/erc20/keeper"
	evmkeeper "github.com/evmos/evmos/v14/x/evm/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v16
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ek *evmkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// Instantiate the (W)ERC20 Precompile for each registered IBC Coin

		// IMPORTANT (@fedekunze): This logic needs to be included on EVERY UPGRADE
		// from now on because the AvailablePrecompiles function does not have access
		// to the state (in this case, the registered token pairs).
		if err := erc20Keeper.RegisterERC20Extensions(ctx); err != nil {
			logger.Error("failed to register ERC-20 Extensions", "error", err.Error())
		}

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
