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

// IterateFees iterates over all registered `FeeContracts` and performs a
// callback.
func (k Keeper) IterateFees(
	ctx sdk.Context,
	handlerFn func(fee types.FeeContract) (stop bool),
) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixFee)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var fee types.FeeContract
		k.cdc.MustUnmarshal(iterator.Value(), &fee)

		if handlerFn(fee) {
			break
		}
	}
}

// GetFee - get registered fee from the identifier
func (k Keeper) GetFee(
	ctx sdk.Context,
	contract common.Address,
) (types.FeeContract, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	bz := store.Get(contract.Bytes())
	if len(bz) == 0 {
		return types.FeeContract{}, false
	}

	var fee types.FeeContract
	k.cdc.MustUnmarshal(bz, &fee)
	return fee, true
}

// SetFee stores a fee
func (k Keeper) SetFee(ctx sdk.Context, fee types.FeeContract) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	key := common.HexToAddress(fee.Contract)
	bz := k.cdc.MustMarshal(&fee)
	store.Set(key.Bytes(), bz)
}

// DeleteFee removes a fee
func (k Keeper) DeleteFee(ctx sdk.Context, fee types.FeeContract) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	key := common.HexToAddress(fee.Contract)
	store.Delete(key.Bytes())
}

// IsFeeRegistered - check if registered FeeContract is registered
func (k Keeper) IsFeeRegistered(
	ctx sdk.Context,
	contract common.Address,
) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	return store.Has(contract.Bytes())
}
