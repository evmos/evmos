package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/inflation/types"
)

// GetPeriod gets current period
func (k Keeper) GetPeriod(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyPrefixPeriod)
	if bz == nil {
		return 0
	}

	return int64(sdk.BigEndianToUint64(bz))
}

// SetPeriod stores the current period
func (k Keeper) SetPeriod(ctx sdk.Context, period int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyPrefixPeriod, sdk.Uint64ToBigEndian(uint64(period)))
}
