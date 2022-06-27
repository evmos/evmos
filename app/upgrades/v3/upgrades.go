package v3

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"

	feemarketv010 "github.com/evmos/ethermint/x/feemarket/migrations/v010"
	feemarketv09types "github.com/evmos/ethermint/x/feemarket/migrations/v09/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
)

// MigrateGenesis migrates exported state from v2 to v3 genesis state.
// It performs a no-op if the migration errors.
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

	return appState
}
