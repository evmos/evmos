// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v19

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	evmkeeper "github.com/evmos/evmos/v18/x/evm/keeper"
)

const (
	StrideOutpostAddress  = "0x0000000000000000000000000000000000000900"
	OsmosisOutpostAddress = "0x0000000000000000000000000000000000000901"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v19
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	evmKeeper *evmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)
		// revenue module is deprecated
		logger.Debug("deleting revenue module from version map...")
		delete(vm, "revenue")

		ctxCache, writeFn := ctx.CacheContext()
		if err := RemoveOutpostsFromEvmParams(ctxCache, evmKeeper); err == nil {
			writeFn()
		}

		// Leave modules as-is to avoid running InitGenesis.
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func RemoveOutpostsFromEvmParams(ctx sdk.Context,
	evmKeeper *evmkeeper.Keeper,
) error {
	params := evmKeeper.GetParams(ctx)
	newActivePrecompiles := make([]string, 0)
	for _, precompile := range params.ActivePrecompiles {
		if precompile != OsmosisOutpostAddress &&
			precompile != StrideOutpostAddress {
			newActivePrecompiles = append(newActivePrecompiles, precompile)
		}
	}
	params.ActivePrecompiles = newActivePrecompiles
	return evmKeeper.SetParams(ctx, params)
}
