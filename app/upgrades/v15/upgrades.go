// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v15

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	p256 "github.com/evmos/evmos/v14/precompiles/p256"
	"github.com/evmos/evmos/v14/utils"
	evmkeeper "github.com/evmos/evmos/v14/x/evm/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v15.0.0
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	evmKeeper *evmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// enable secp256r1 precompile on testnet
		if utils.IsTestnet(ctx.ChainID()) {
			if err := EnableP256Precompile(ctx, evmKeeper); err != nil {
				logger.Error("failed to enable secp256r1 precompile", "error", err.Error())
			}
		}

		// we are deprecating crisis module since it is not being used
		logger.Debug("deleting crisis module from version map...")
		delete(vm, "crisis")

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// EnableP256Precompile appends the address of the P256 Precompile
// to the list of active precompiles.
func EnableP256Precompile(ctx sdk.Context, evmKeeper *evmkeeper.Keeper) error {
	// Get the list of active precompiles from the genesis state
	params := evmKeeper.GetParams(ctx)
	activePrecompiles := params.ActivePrecompiles
	activePrecompiles = append(activePrecompiles, p256.Precompile{}.Address().String())
	params.ActivePrecompiles = activePrecompiles

	return evmKeeper.SetParams(ctx, params)
}
