package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/evmos/x/incentives/types"
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
	var incentive types.Incentive
	bz := store.Get(contract.Bytes())
	if len(bz) == 0 {
		return types.Incentive{}, false
	}

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

// DeleteIncentive removes an incentive.
func (k Keeper) DeleteIncentive(ctx sdk.Context, incentive types.Incentive) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixIncentive)
	key := common.HexToAddress(incentive.Contract)
	store.Delete(key.Bytes())
}

// IsIncentiveRegistered - check if registered Incentive is registered
func (k Keeper) IsIncentiveRegistered(
	ctx sdk.Context,
	contract common.Address,
) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixIncentive)
	return store.Has(contract.Bytes())
}

// GetIncentiveTotalGas - Get total cummulative gas of a given incentive
func (k Keeper) GetIncentiveTotalGas(
	ctx sdk.Context,
	incentive types.Incentive,
) uint64 {
	in, _ := k.GetIncentive(ctx, common.HexToAddress(incentive.Contract))
	return in.TotalGas
}

// Set total cummulative gas of a given incentive
func (k Keeper) SetIncentiveTotalGas(
	ctx sdk.Context,
	incentive types.Incentive,
	gas uint64,
) {
	in, _ := k.GetIncentive(ctx, common.HexToAddress(incentive.Contract))
	in.TotalGas = gas
	k.SetIncentive(ctx, in)
}
