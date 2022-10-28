package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v9/x/incentives/types"
)

// GetAllIncentives - get all registered Incentives
func (k Keeper) GetAllIncentives(ctx sdk.Context) []types.Incentive {
	incentives := []types.Incentive{}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixIncentive)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var incentive types.Incentive
		k.cdc.MustUnmarshal(iterator.Value(), &incentive)

		incentives = append(incentives, incentive)
	}

	return incentives
}

// IterateIncentives iterates over all registered `Incentives` and performs a
// callback.
func (k Keeper) IterateIncentives(
	ctx sdk.Context,
	handlerFn func(incentive types.Incentive) (stop bool),
) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixIncentive)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var incentive types.Incentive
		k.cdc.MustUnmarshal(iterator.Value(), &incentive)

		if handlerFn(incentive) {
			break
		}
	}
}

// GetIncentive - get registered incentive from the identifier
func (k Keeper) GetIncentive(
	ctx sdk.Context,
	contract common.Address,
) (types.Incentive, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixIncentive)
	bz := store.Get(contract.Bytes())
	if len(bz) == 0 {
		return types.Incentive{}, false
	}

	var incentive types.Incentive
	k.cdc.MustUnmarshal(bz, &incentive)
	return incentive, true
}

// SetIncentive stores an incentive
func (k Keeper) SetIncentive(ctx sdk.Context, incentive types.Incentive) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixIncentive)
	key := common.HexToAddress(incentive.Contract)
	bz := k.cdc.MustMarshal(&incentive)
	store.Set(key.Bytes(), bz)
}

// DeleteIncentiveAndUpdateAllocationMeters removes an incentive and updates the
// percentage of incentives allocated to each denomination.
func (k Keeper) DeleteIncentiveAndUpdateAllocationMeters(ctx sdk.Context, incentive types.Incentive) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixIncentive)
	key := common.HexToAddress(incentive.Contract)
	store.Delete(key.Bytes())

	// Subtract allocations from allocation meters
	for _, al := range incentive.Allocations {
		// NOTE: existence of incentive is already checked
		am, _ := k.GetAllocationMeter(ctx, al.Denom)
		amount := am.Amount.Sub(al.Amount)
		am = sdk.DecCoin{
			Denom:  al.Denom,
			Amount: amount,
		}

		k.SetAllocationMeter(ctx, am)
	}
}

// IsIncentiveRegistered - check if registered Incentive is registered
func (k Keeper) IsIncentiveRegistered(
	ctx sdk.Context,
	contract common.Address,
) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixIncentive)
	return store.Has(contract.Bytes())
}

// Set total cummulative gas of a given incentive
func (k Keeper) SetIncentiveTotalGas(
	ctx sdk.Context,
	incentive types.Incentive,
	gas uint64,
) {
	incentive.TotalGas = gas
	k.SetIncentive(ctx, incentive)
}
