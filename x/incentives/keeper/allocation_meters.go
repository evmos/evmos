package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/incentives/types"
)

// GetAllAllocationMeters - get all registered AllocationMeters
func (k Keeper) GetAllAllocationMeters(ctx sdk.Context) []types.AllocationMeter {
	allocationMeters := []types.AllocationMeter{}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixAllocationMeter)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var allocationMeter types.AllocationMeter
		k.cdc.MustUnmarshal(iterator.Value(), &allocationMeter)

		allocationMeters = append(allocationMeters, allocationMeter)
	}

	return allocationMeters
}

// GetAllocationMeter - get registered allocationMeter from the identifier
func (k Keeper) GetAllocationMeter(
	ctx sdk.Context,
	denom string,
) (sdk.Dec, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllocationMeter)
	var allocationMeter types.AllocationMeter
	bz := store.Get([]byte(denom))
	if len(bz) == 0 {
		return sdk.Dec{}, false
	}

	k.cdc.MustUnmarshal(bz, &allocationMeter)
	return allocationMeter.Allocation.Amount, true
}

// SetAllocationMeter stores an allocationMeter
func (k Keeper) SetAllocationMeter(ctx sdk.Context, am types.AllocationMeter) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllocationMeter)
	key := []byte(am.Allocation.Denom)
	intBz, _ := am.Allocation.Amount.Marshal()
	store.Set(key, intBz)
}

// DeleteAllocationMeter removes an allocationMeter.
func (k Keeper) DeleteAllocationMeter(ctx sdk.Context, am types.AllocationMeter) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllocationMeter)
	key := []byte(am.Allocation.Denom)
	store.Delete(key)
}
