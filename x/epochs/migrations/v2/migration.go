package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v1 "github.com/tharsis/evmos/v4/x/epochs/migrations/v1"
	"github.com/tharsis/evmos/v4/x/epochs/types"
)

// MigrateStore migrates in-place store migrations from v1 to v2. The migration
// orders epochs numerically (ascending) after their duration. Changes include:
// - Change epoch info key to be its duration
func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.Codec) error {
	store := ctx.KVStore(storeKey)

	oldStore := prefix.NewStore(store, types.KeyPrefixEpoch)
	epochStore := oldStore

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		epochInfo := types.EpochInfo{}
		cdc.MustUnmarshal(oldStoreIter.Value(), &epochInfo)
		duration := types.DurationToBz(epochInfo.Duration)

		// Set epoch info by new duration key.
		epochStore.Set(duration, oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}
	return nil
}

// MigrateJSON accepts exported 1 x/epochs genesis state and migrates it
// to 2 x/epochs genesis state. Identifiers are removed.
func MigrateJSON(oldState v1.GenesisState) types.GenesisState {
	var newState types.GenesisState

	for _, epoch := range oldState.Epochs {
		newepoch := types.EpochInfo{
			StartTime:               epoch.StartTime,
			Duration:                epoch.Duration,
			CurrentEpoch:            epoch.CurrentEpoch,
			CurrentEpochStartTime:   epoch.CurrentEpochStartTime,
			EpochCountingStarted:    epoch.EpochCountingStarted,
			CurrentEpochStartHeight: epoch.CurrentEpochStartHeight,
		}
		newState.Epochs = append(newState.Epochs, newepoch)
	}
	return newState
}
