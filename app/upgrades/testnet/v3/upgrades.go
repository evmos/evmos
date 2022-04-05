package v3

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	feemarketv010 "github.com/tharsis/ethermint/x/feemarket/migrations/v010"
	feemarketv09types "github.com/tharsis/ethermint/x/feemarket/migrations/v09/types"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"
	v1claims "github.com/tharsis/evmos/v3/x/claims/migrations/v1/types"
	v2claims "github.com/tharsis/evmos/v3/x/claims/migrations/v2"
	claimstypes "github.com/tharsis/evmos/v3/x/claims/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v3
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Refs:
		// - https://docs.cosmos.network/master/building-modules/upgrade.html#registering-migrations
		// - https://docs.cosmos.network/master/migrations/chain-upgrade-guide-044.html#chain-upgrade

		// migrate fee market module and the claims module
		// avoid running InitGenesis.
		vm[feemarkettypes.ModuleName] = 1
		vm[claimstypes.ModuleName] = 1

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// MigrateGenesis migrates exported state from v2 to v3 genesis state. It performs a no-op if the migration errors.
func MigrateGenesis(appState types.AppMap, clientCtx client.Context) types.AppMap {
	// Migrate x/feemarket.
	if appState[feemarkettypes.ModuleName] == nil {
		return appState
	}

	// unmarshal relative source genesis application state
	var oldFeeMarketState feemarketv09types.GenesisState
	if err := clientCtx.Codec.UnmarshalJSON(appState[feemarkettypes.ModuleName], &oldFeeMarketState); err != nil {
		return appState
	}

	// delete deprecated x/feemarket genesis state
	delete(appState, feemarkettypes.ModuleName)

	// Migrate relative source genesis application state and marshal it into
	// the respective key.
	newFeeMarketState := feemarketv010.MigrateJSON(oldFeeMarketState)

	feeMarketBz, err := clientCtx.Codec.MarshalJSON(&newFeeMarketState)
	if err != nil {
		return appState
	}

	appState[feemarkettypes.ModuleName] = feeMarketBz

	// unmarshal relative source genesis application state
	var oldClaimsState v1claims.GenesisState
	if err := clientCtx.Codec.UnmarshalJSON(appState[claimstypes.ModuleName], &oldClaimsState); err != nil {
		return appState
	}

	// delete deprecated x/feemarket genesis state
	delete(appState, claimstypes.ModuleName)

	// Migrate relative source genesis application state and marshal it into
	// the respective key.
	newClaimsState := v2claims.MigrateJSON(oldClaimsState)

	claimsBz, err := clientCtx.Codec.MarshalJSON(&newClaimsState)
	if err != nil {
		return appState
	}

	appState[claimstypes.ModuleName] = claimsBz

	return appState
}
