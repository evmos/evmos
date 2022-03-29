package v3

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"
)

const UpgradeName = "v3"

// CreateUpgradeHandler creates an SDK upgrade handler for v3
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Refs:
		// - https://docs.cosmos.network/master/building-modules/upgrade.html#registering-migrations
		// - https://docs.cosmos.network/master/migrations/chain-upgrade-guide-044.html#chain-upgrade

		// migrate fee market module, other modules are left as-is to
		// avoid running InitGenesis.
		vm[feemarkettypes.ModuleName] = 1

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// Migrate migrates exported state from v2 to v3 genesis state.
func Migrate(appState types.AppMap, clientCtx client.Context) types.AppMap {
	// Migrate x/feemarket.
	if appState[feemarkettypes.ModuleName] == nil {
		return appState
	}

	// unmarshal relative source genesis application state
	var oldFeeMarketState feemarkettypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appState[feemarkettypes.ModuleName], &oldFeeMarketState)

	// delete deprecated x/feemarket genesis state
	delete(appState, feemarkettypes.ModuleName)

	// Migrate relative source genesis application state and marshal it into
	// the respective key.
	newFeeMarketState := MigrateJSON(oldFeeMarketState)

	appState[feemarkettypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&newFeeMarketState)

	return appState
}

// MigrateJSON accepts exported v2 x/feemarket genesis state and migrates it to
// v3 x/feemarket genesis state. The migration includes:
// - Migrate BaseFee to params
func MigrateJSON(oldState feemarkettypes.GenesisState) feemarkettypes.GenesisState {
	return feemarkettypes.GenesisState{
		Params: feemarkettypes.Params{
			NoBaseFee:                oldState.Params.NoBaseFee,
			BaseFeeChangeDenominator: oldState.Params.BaseFeeChangeDenominator,
			ElasticityMultiplier:     oldState.Params.ElasticityMultiplier,
			EnableHeight:             oldState.Params.EnableHeight,
			// BaseFee:                  oldState.BaseFee, FIXME: import
		},
		BlockGas: oldState.BlockGas,
	}
}
