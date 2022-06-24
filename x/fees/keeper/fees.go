package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v5/x/fees/types"
)

// GetFees - get all registered Fees
func (k Keeper) GetFees(ctx sdk.Context) []types.Fee {
	fees := []types.Fee{}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixFee)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var fee types.Fee
		k.cdc.MustUnmarshal(iterator.Value(), &fee)

		fees = append(fees, fee)
	}

	return fees
}

// IterateFees iterates over all registered contracts and performs a
// callback with the corresponding Fee.
func (k Keeper) IterateFees(
	ctx sdk.Context,
	handlerFn func(fee types.Fee) (stop bool),
) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixFee)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var fee types.Fee
		k.cdc.MustUnmarshal(iterator.Value(), &fee)

		if handlerFn(fee) {
			break
		}
	}
}

// GetFee returns Fee for a registered contract
func (k Keeper) GetFee(
	ctx sdk.Context,
	contract common.Address,
) (types.Fee, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	bz := store.Get(contract.Bytes())
	if len(bz) == 0 {
		return types.Fee{}, false
	}

	var fee types.Fee
	k.cdc.MustUnmarshal(bz, &fee)
	return fee, true
}

// SetFee stores the Fee for a registered contract.
func (k Keeper) SetFee(ctx sdk.Context, fee types.Fee) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	key := common.HexToAddress(fee.ContractAddress)
	bz := k.cdc.MustMarshal(&fee)
	store.Set(key.Bytes(), bz)
}

// DeleteFee deletes a fee contract
func (k Keeper) DeleteFee(ctx sdk.Context, fee types.Fee) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	key := common.HexToAddress(fee.ContractAddress)
	store.Delete(key.Bytes())
}

// SetDeployerMap stores a fee contract by deployer mapping
func (k Keeper) SetDeployerMap(
	ctx sdk.Context,
	deployer sdk.AccAddress,
	contract common.Address,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDeployer)
	key := append(deployer.Bytes(), contract.Bytes()...)
	store.Set(key, []byte{1})
}

// DeleteDeployerMap deletes a fee contract by deployer mapping
func (k Keeper) DeleteDeployerMap(
	ctx sdk.Context,
	deployer sdk.AccAddress,
	contract common.Address,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDeployer)
	key := append(deployer.Bytes(), contract.Bytes()...)
	store.Delete(key)
}

// SetWithdrawMap stores a fee contract by withdraw address mapping
func (k Keeper) SetWithdrawMap(
	ctx sdk.Context,
	withdraw sdk.AccAddress,
	contract common.Address,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixWithdraw)
	key := append(withdraw.Bytes(), contract.Bytes()...)
	store.Set(key, []byte{1})
}

// DeleteWithdrawMap deletes a fee contract by withdraw address mapping
func (k Keeper) DeleteWithdrawMap(
	ctx sdk.Context,
	withdraw sdk.AccAddress,
	contract common.Address,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixWithdraw)
	key := append(withdraw.Bytes(), contract.Bytes()...)
	store.Delete(key)
}

// IsFeeRegistered checks if a contract was registered for receiving fees
func (k Keeper) IsFeeRegistered(
	ctx sdk.Context,
	contract common.Address,
) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	return store.Has(contract.Bytes())
}

// IsDeployerMapSet checks if a fee contract by deployer address mapping is set
// in store
func (k Keeper) IsDeployerMapSet(
	ctx sdk.Context,
	deployer sdk.AccAddress,
	contract common.Address,
) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDeployer)
	key := append(deployer.Bytes(), contract.Bytes()...)
	return store.Has(key)
}

// IsWithdrawMapSet checks if a fee contract by withdraw address mapping is set
// in store
func (k Keeper) IsWithdrawMapSet(
	ctx sdk.Context,
	withdraw sdk.AccAddress,
	contract common.Address,
) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixWithdraw)
	key := append(withdraw.Bytes(), contract.Bytes()...)
	return store.Has(key)
}
