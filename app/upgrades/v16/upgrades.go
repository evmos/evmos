// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v16

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/evmos/evmos/v15/precompiles/p256"
	"github.com/evmos/evmos/v15/utils"
	evmkeeper "github.com/evmos/evmos/v15/x/evm/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v16.0.0
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ek *evmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		p256Address := p256.Precompile{}.Address()
		// enable secp256r1 precompile on testnet
		if utils.IsTestnet(ctx.ChainID()) {
			if err := ek.EnablePrecompiles(ctx, p256Address); err != nil {
				logger.Error("failed to enable precompiles", "error", err.Error())
			}
		}

		// recovery module is deprecated since it is renamed to "revenue" module
		logger.Debug("deleting recovery module from version map...")
		delete(vm, "recovery")

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
