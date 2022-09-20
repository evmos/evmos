package keeper

import (
	"github.com/ArableProtocol/acrechain/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetNextReductionTime returns next reduction time.
func (k Keeper) GetNextReductionTime(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.LastReductionTimeKey)
	if b == nil {
		return 0
	}

	return int64(sdk.BigEndianToUint64(b))
}

// SetNextReductionTime set next reduction time.
func (k Keeper) SetNextReductionTime(ctx sdk.Context, time int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastReductionTimeKey, sdk.Uint64ToBigEndian(uint64(time)))
}

// get the minter.
func (k Keeper) GetMinter(ctx sdk.Context) (minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.MinterKey)
	if b == nil {
		panic("stored minter should not have been nil")
	}

	k.cdc.MustUnmarshal(b, &minter)
	return
}

// set the minter.
func (k Keeper) SetMinter(ctx sdk.Context, minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&minter)
	store.Set(types.MinterKey, b)
}
