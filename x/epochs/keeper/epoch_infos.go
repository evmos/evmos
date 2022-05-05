package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/evmos/v4/x/epochs/types"
)

// GetEpochInfo returns epoch info by identifier
func (k Keeper) GetEpochInfo(ctx sdk.Context, identifier string) (types.EpochInfo, bool) {
	duration, found := k.GetEpochDuration(ctx, identifier)
	if !found {
		return types.EpochInfo{}, false
	}
	return k.GetEpoch(ctx, duration)
}

// GetEpochDuration returns epoch duration by identifier
func (k Keeper) GetEpochDuration(ctx sdk.Context, identifier string) ([]byte, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEpochDuration)
	bz := store.Get([]byte(identifier))
	if len(bz) == 0 {
		return make([]byte, 0), false
	}
	return bz, true
}

// GetEpochInfo returns epoch info by duration
func (k Keeper) GetEpoch(ctx sdk.Context, duration []byte) (types.EpochInfo, bool) {
	epoch := types.EpochInfo{}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEpoch)
	bz := store.Get(duration)
	if len(bz) == 0 {
		return epoch, false
	}

	k.cdc.MustUnmarshal(bz, &epoch)
	return epoch, true
}

// SetEpochInfo set epoch info
func (k Keeper) SetEpochInfo(ctx sdk.Context, epoch types.EpochInfo) {
	k.setEpochDuration(ctx, epoch)
	k.setEpoch(ctx, epoch)
}

// SetEpochDuration set epoch duration by identifier
func (k Keeper) setEpochDuration(ctx sdk.Context, epoch types.EpochInfo) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEpochDuration)
	store.Set([]byte(epoch.Identifier), durationToBz(epoch.Duration))
}

// SetEpochInfo set epoch duration by identifier
func (k Keeper) setEpoch(ctx sdk.Context, epoch types.EpochInfo) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEpoch)
	bz := k.cdc.MustMarshal(&epoch)
	store.Set(durationToBz(epoch.Duration), bz)
}

// DeleteEpochInfo delete epoch info
func (k Keeper) DeleteEpochInfo(ctx sdk.Context, identifier string) {
	duration, found := k.GetEpochDuration(ctx, identifier)
	if found {
		k.deleteEpochDuration(ctx, identifier)
		k.deleteEpoch(ctx, duration)
	}
}

// DeleteEpochDuration delete epoch duration by identifier
func (k Keeper) deleteEpochDuration(ctx sdk.Context, identifier string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEpochDuration)
	store.Delete([]byte(identifier))
}

// DeleteEpoch delete epoch info
func (k Keeper) deleteEpoch(ctx sdk.Context, duration []byte) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEpoch)
	store.Delete(duration)
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

// durationToBz parses time duration to maintain number-compatible ordering
func durationToBz(duration time.Duration) []byte {
	// 13 digits left padded with zero, allows for 300 year durations
	s := fmt.Sprintf("%013d", duration.Milliseconds())
	return []byte(s)
}
