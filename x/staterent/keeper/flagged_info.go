// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/x/staterent/types"
)

// GetState loads contract state from database.
func (k *Keeper) GetFlaggedInfo(ctx sdk.Context, addr common.Address) *types.FlaggedInfo {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFlaggedInfo)
	value := store.Get(addr.Bytes())
	if len(value) == 0 {
		return nil
	}
	var v types.FlaggedInfo
	if err := k.cdc.Unmarshal(value, &v); err != nil {
		// TODO: should we panic here?
		return nil
	}
	return &v
}

// SetState update contract storage.
func (k *Keeper) SetFlaggedInfo(ctx sdk.Context, addr common.Address, value types.FlaggedInfo) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFlaggedInfo)
	bz, err := k.cdc.Marshal(&value)
	if err != nil {
		// TODO: should we panic here?
		return
	}

	store.Set(addr.Bytes(), bz)

	k.Logger(ctx).Debug(
		"flagged contract",
		"address", addr.Hex(),
	)
}

// DeleteFlaggedInfo deletes an entry from the flagged info array
func (k *Keeper) DeleteFlaggedInfo(ctx sdk.Context, addr common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFlaggedInfo)
	store.Delete(addr.Bytes())

	k.Logger(ctx).Debug(
		"removed flag from contract ",
		"address", addr.Hex(),
	)
}

// IterateFlaggedInfo iterate through all flagged contracts
func (k Keeper) IterateFlaggedInfo(ctx sdk.Context, fn func(index int64, info types.FlaggedInfo) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFlaggedInfo)
	iterator := sdk.KVStorePrefixIterator(store, nil)
	defer iterator.Close()

	i := int64(0)

	for ; iterator.Valid(); iterator.Next() {
		info := types.FlaggedInfo{}
		k.cdc.MustUnmarshal(iterator.Value(), &info)

		stop := fn(i, info)

		if stop {
			break
		}
		i++
	}
}
