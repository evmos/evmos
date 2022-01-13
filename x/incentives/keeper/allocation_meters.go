package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/incentives/types"
)

// GetAllAllocationMeters - get all registered AllocationMeters
func (k Keeper) GetAllAllocationMeters(ctx sdk.Context) []sdk.DecCoin {
	allocationMeters := []sdk.DecCoin{}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixAllocationMeter)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {

		denom := string(iterator.Value())
		allocationMeter, found := k.GetAllocationMeter(ctx, denom)
		if !found {
			panic(fmt.Errorf("unable to unmarshal amount value %v", found))
		}

		allocationMeters = append(allocationMeters, allocationMeter)
	}

	return allocationMeters
}

// GetAllocationMeter - get registered allocationMeter from the identifier
func (k Keeper) GetAllocationMeter(
	ctx sdk.Context,
	denom string,
) (sdk.DecCoin, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllocationMeter)

	bz := store.Get([]byte(denom))
	if bz == nil {
		return sdk.DecCoin{
			Denom:  denom,
			Amount: sdk.ZeroDec(),
		}, false
	}

	var amount sdk.Dec
	err := amount.Unmarshal(bz)
	if err != nil {
		panic(fmt.Errorf("unable to unmarshal amount value %v", err))
	}
	return sdk.DecCoin{
		Denom:  denom,
		Amount: amount,
	}, true
}

// SetAllocationMeter stores an allocationMeter
func (k Keeper) SetAllocationMeter(ctx sdk.Context, am sdk.DecCoin) {
	bz, err := am.Amount.Marshal()
	if err != nil {
		panic(fmt.Errorf("unable to marshal amount value %v", err))
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllocationMeter)
	key := []byte(am.Denom)
	store.Set(key, bz)
}

// DeleteAllocationMeter removes an allocationMeter.
func (k Keeper) DeleteAllocationMeter(ctx sdk.Context, am sdk.DecCoin) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllocationMeter)
	key := []byte(am.Denom)
	store.Delete(key)
}
