// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v16

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	osmosisoutpost "github.com/evmos/evmos/v16/precompiles/outposts/osmosis"
	strideoutpost "github.com/evmos/evmos/v16/precompiles/outposts/stride"
	"github.com/evmos/evmos/v16/precompiles/p256"
	"github.com/evmos/evmos/v16/utils"
	evmkeeper "github.com/evmos/evmos/v16/x/evm/keeper"
	inflationkeeper "github.com/evmos/evmos/v16/x/inflation/v1/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v16.0.0
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ek *evmkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	inflationKeeper inflationkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// enable secp256r1 precompile on testnet
		if utils.IsTestnet(ctx.ChainID()) {
			p256Address := p256.Precompile{}.Address()

			if err := ek.EnablePrecompiles(ctx, p256Address); err != nil {
				logger.Error("failed to enable precompiles", "error", err.Error())
			}
		}

		// enable stride and osmosis outposts
		strideAddress := strideoutpost.Precompile{}.Address()
		osmosisAddress := osmosisoutpost.Precompile{}.Address()
		if err := ek.EnablePrecompiles(ctx, strideAddress, osmosisAddress); err != nil {
			logger.Error("failed to enable outposts", "error", err.Error())
		}

		if err := BurnUsageIncentivesPool(ctx, bankKeeper); err != nil {
			logger.Error("failed to burn inflation pool", "error", err.Error())
		}

		if err := UpdateInflationParams(ctx, inflationKeeper); err != nil {
			logger.Error("failed to update inflation params", "error", err.Error())
		}

		// recovery module is deprecated
		logger.Debug("deleting recovery module from version map...")
		delete(vm, "recovery")
		logger.Debug("deleting claims module from version map...")
		delete(vm, "claims")
		logger.Debug("deleting incentives module from version map...")
		delete(vm, "incentives")

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
