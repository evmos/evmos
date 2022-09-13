package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v9/x/feesplit/types"
)

// GetFeeSplits returns all registered FeeSplits.
func (k Keeper) GetFeeSplits(ctx sdk.Context) []types.FeeSplit {
	feeSplits := []types.FeeSplit{}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixFeeSplit)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var feeSplit types.FeeSplit
		k.cdc.MustUnmarshal(iterator.Value(), &feeSplit)

		feeSplits = append(feeSplits, feeSplit)
	}

	return feeSplits
}

// IterateFeeSplits iterates over all registered contracts and performs a
// callback with the corresponding FeeSplit.
func (k Keeper) IterateFeeSplits(
	ctx sdk.Context,
	handlerFn func(fee types.FeeSplit) (stop bool),
) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixFeeSplit)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var feeSplit types.FeeSplit
		k.cdc.MustUnmarshal(iterator.Value(), &feeSplit)

		if handlerFn(feeSplit) {
			break
		}
	}
}

// GetFeeSplit returns the FeeSplit for a registered contract
func (k Keeper) GetFeeSplit(
	ctx sdk.Context,
	contract common.Address,
) (types.FeeSplit, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFeeSplit)
	bz := store.Get(contract.Bytes())
	if len(bz) == 0 {
		return types.FeeSplit{}, false
	}

	var feeSplit types.FeeSplit
	k.cdc.MustUnmarshal(bz, &feeSplit)
	return feeSplit, true
}

// SetFeeSplit stores the FeeSplit for a registered contract.
func (k Keeper) SetFeeSplit(ctx sdk.Context, feeSplit types.FeeSplit) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFeeSplit)
	key := feeSplit.GetContractAddr()
	bz := k.cdc.MustMarshal(&feeSplit)
	store.Set(key.Bytes(), bz)
}

// DeleteFeeSplit deletes a FeeSplit of a registered contract.
func (k Keeper) DeleteFeeSplit(ctx sdk.Context, fee types.FeeSplit) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFeeSplit)
	key := fee.GetContractAddr()
	store.Delete(key.Bytes())
}

// SetDeployerMap stores a contract-by-deployer mapping
func (k Keeper) SetDeployerMap(
	ctx sdk.Context,
	deployer sdk.AccAddress,
	contract common.Address,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDeployer)
	key := append(deployer.Bytes(), contract.Bytes()...)
	store.Set(key, []byte{1})
}

// DeleteDeployerMap deletes a contract-by-deployer mapping
func (k Keeper) DeleteDeployerMap(
	ctx sdk.Context,
	deployer sdk.AccAddress,
	contract common.Address,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDeployer)
	key := append(deployer.Bytes(), contract.Bytes()...)
	store.Delete(key)
}

// SetWithdrawerMap stores a contract-by-withdrawer mapping
func (k Keeper) SetWithdrawerMap(
	ctx sdk.Context,
	withdrawer sdk.AccAddress,
	contract common.Address,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixWithdrawer)
	key := append(withdrawer.Bytes(), contract.Bytes()...)
	store.Set(key, []byte{1})
}

// DeleteWithdrawMap deletes a contract-by-withdrawer mapping
func (k Keeper) DeleteWithdrawerMap(
	ctx sdk.Context,
	withdrawer sdk.AccAddress,
	contract common.Address,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixWithdrawer)
	key := append(withdrawer.Bytes(), contract.Bytes()...)
	store.Delete(key)
}

// IsFeeSplitRegistered checks if a contract was registered for receiving
// transaction fees
func (k Keeper) IsFeeSplitRegistered(
	ctx sdk.Context,
	contract common.Address,
) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFeeSplit)
	return store.Has(contract.Bytes())
}

// IsDeployerMapSet checks if a given contract-by-withdrawer mapping is set in
// store
func (k Keeper) IsDeployerMapSet(
	ctx sdk.Context,
	deployer sdk.AccAddress,
	contract common.Address,
) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixDeployer)
	key := append(deployer.Bytes(), contract.Bytes()...)
	return store.Has(key)
}

// IsWithdrawerMapSet checks if a giveb contract-by-withdrawer mapping is set in
// store
func (k Keeper) IsWithdrawerMapSet(
	ctx sdk.Context,
	withdrawer sdk.AccAddress,
	contract common.Address,
) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixWithdrawer)
	key := append(withdrawer.Bytes(), contract.Bytes()...)
	return store.Has(key)
}
