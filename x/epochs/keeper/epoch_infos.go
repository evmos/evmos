// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v10/x/epochs/types"
)

// GetEpochInfo returns epoch info by identifier
func (k Keeper) GetEpochInfo(ctx sdk.Context, identifier string) (types.EpochInfo, bool) {
	epoch := types.EpochInfo{}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEpoch)
	bz := store.Get([]byte(identifier))
	if len(bz) == 0 {
		return epoch, false
	}

	k.cdc.MustUnmarshal(bz, &epoch)
	return epoch, true
}

// SetEpochInfo set epoch info
func (k Keeper) SetEpochInfo(ctx sdk.Context, epoch types.EpochInfo) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEpoch)
	bz := k.cdc.MustMarshal(&epoch)
	store.Set([]byte(epoch.Identifier), bz)
}

// DeleteEpochInfo delete epoch info
func (k Keeper) DeleteEpochInfo(ctx sdk.Context, identifier string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEpoch)
	store.Delete([]byte(identifier))
}

// IterateEpochInfo iterate through epochs
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
