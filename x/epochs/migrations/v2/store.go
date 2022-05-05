package v2

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/evmos/v4/x/epochs/types"
)

// MigrateStore migrates in-place store migrations from v1 to v2. The migration
// orders epochs numerically (ascending) after their duration. Changes include:
// - Change epoch info key to be its duration
// - Add another KVStore for storing the duration, using the epoch identifier as key
func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.Codec) error {
	store := ctx.KVStore(storeKey)

	oldStore := prefix.NewStore(store, KeyPrefixEpoch)
	durationStore := prefix.NewStore(store, KeyPrefixEpochDuration)
	epochStore := oldStore

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		epochInfo := types.EpochInfo{}
		cdc.MustUnmarshal(oldStoreIter.Value(), &epochInfo)
		duration := DurationToBz(epochInfo.Duration)

		// Set epoch duration by identifier in place of the old epoch info
		durationStore.Set(oldStoreIter.Key(), duration)

		// Set epoch info by new duration key. Values don't change.
		epochStore.Set(duration, oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}
	return nil
}

// durationToBz parses time duration to maintain number-compatible ordering
func DurationToBz(duration time.Duration) []byte {
	// 13 digits left padded with zero, allows for 300 year durations
	s := fmt.Sprintf("%013d", duration.Milliseconds())
	return []byte(s)
}
