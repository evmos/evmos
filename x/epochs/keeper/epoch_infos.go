package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/evmos/v4/x/epochs/types"
)

// GetEpochInfo returns epoch info by identifier
func (k Keeper) GetEpochInfo(ctx sdk.Context, identifier string) (types.EpochInfo, bool) {
	epoch := types.EpochInfo{}
	duration, exists := types.IdentifierToDuration[identifier]
	if !exists {
		return epoch, false
	}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEpoch)
	bz := store.Get(types.DurationToBz(duration))
	if len(bz) == 0 {
		return epoch, false
	}

	k.cdc.MustUnmarshal(bz, &epoch)
	return epoch, true
}

// SetEpochInfo set epoch duration by identifier
func (k Keeper) SetEpochInfo(ctx sdk.Context, epoch types.EpochInfo) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEpoch)
	bz := k.cdc.MustMarshal(&epoch)
	store.Set(types.DurationToBz(epoch.Duration), bz)
}

// DeleteEpochInfo delete epoch info
func (k Keeper) DeleteEpochInfo(ctx sdk.Context, duration time.Duration) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEpoch)
	store.Delete(types.DurationToBz(duration))
}

// IterateEpochInfo iterate through epochs in ascending numerical order, by duration
func (k Keeper) IterateEpochInfo(ctx sdk.Context, fn func(index int64, epochInfo types.EpochInfo) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEpoch)

	iterator := sdk.KVStorePrefixIterator(store, nil)
	defer iterator.Close()

	i := int64(0)

	for ; iterator.Valid(); iterator.Next() {
		epoch := types.EpochInfo{}
		k.cdc.MustUnmarshal(iterator.Value(), &epoch)

		stop := fn(i, epoch)

		if stop {
			break
		}
		i++
	}
}

// AllEpochInfos returns every epochInfo in the store
func (k Keeper) AllEpochInfos(ctx sdk.Context) []types.EpochInfo {
	epochs := []types.EpochInfo{}
	k.IterateEpochInfo(ctx, func(_ int64, epochInfo types.EpochInfo) (stop bool) {
		epochs = append(epochs, epochInfo)
		return false
	})
	return epochs
}
