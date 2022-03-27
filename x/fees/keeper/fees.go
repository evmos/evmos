package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/tharsis/evmos/v3/x/fees/types"
)

// GetAllFees - get all registered FeeContract
func (k Keeper) GetAllFees(ctx sdk.Context) []types.FeeContract {
	fees := []types.FeeContract{}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixFee)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var fee types.FeeContract
		k.cdc.MustUnmarshal(iterator.Value(), &fee)

		fees = append(fees, fee)
	}

	return fees
}

// // IterateIncentives iterates over all registered `Incentives` and performs a
// // callback.
// func (k Keeper) IterateIncentives(
// 	ctx sdk.Context,
// 	handlerFn func(incentive types.Incentive) (stop bool),
// ) {
// 	store := ctx.KVStore(k.storeKey)
// 	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixIncentive)
// 	defer iterator.Close()

// 	for ; iterator.Valid(); iterator.Next() {
// 		var incentive types.Incentive
// 		k.cdc.MustUnmarshal(iterator.Value(), &incentive)

// 		if handlerFn(incentive) {
// 			break
// 		}
// 	}
// }

// GetIncentive - get registered incentive from the identifier
func (k Keeper) GetFee(
	ctx sdk.Context,
	contract common.Address,
) (types.FeeContract, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	bz := store.Get(contract.Bytes())
	if len(bz) == 0 {
		return types.FeeContract{}, false
	}

	var incentive types.FeeContract
	k.cdc.MustUnmarshal(bz, &incentive)
	return incentive, true
}

// SetFeeContract stores a fee
func (k Keeper) SetFeeContract(ctx sdk.Context, incentive types.FeeContract) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	key := common.HexToAddress(incentive.Contract)
	bz := k.cdc.MustMarshal(&incentive)
	store.Set(key.Bytes(), bz)
}

// DeleteIncentiveAndUpdateAllocationMeters removes an incentive and updates the
// percentage of incentives allocated to each denomination.
func (k Keeper) DeleteContract(ctx sdk.Context, incentive types.FeeContract) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	key := common.HexToAddress(incentive.Contract)
	store.Delete(key.Bytes())
}

// IsContractRegistered - check if registered Incentive is registered
func (k Keeper) IsContractRegistered(
	ctx sdk.Context,
	contract common.Address,
) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	return store.Has(contract.Bytes())
}
