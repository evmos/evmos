package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/tharsis/evmos/v3/x/fees/types"
)

// GetAllFees - get all registered DevFeeInfo instances
func (k Keeper) GetAllFees(ctx sdk.Context) []types.DevFeeInfo {
	fees := []types.DevFeeInfo{}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixFee)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var fee types.DevFeeInfo
		k.cdc.MustUnmarshal(iterator.Value(), &fee)

		fees = append(fees, fee)
	}

	return fees
}

// IterateFees iterates over all registered `DevFeeInfos` and performs a
// callback.
func (k Keeper) IterateFees(
	ctx sdk.Context,
	handlerFn func(fee types.DevFeeInfo) (stop bool),
) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixFee)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var fee types.DevFeeInfo
		k.cdc.MustUnmarshal(iterator.Value(), &fee)

		if handlerFn(fee) {
			break
		}
	}
}

// GetFee - get registered contract from the identifier
func (k Keeper) GetFee(
	ctx sdk.Context,
	contract common.Address,
) (types.DevFeeInfo, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	bz := store.Get(contract.Bytes())
	if len(bz) == 0 {
		return types.DevFeeInfo{}, false
	}

	var fee types.DevFeeInfo
	k.cdc.MustUnmarshal(bz, &fee)
	return fee, true
}

// SetFee stores a registered contract
func (k Keeper) SetFee(ctx sdk.Context, fee types.DevFeeInfo) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	key := common.HexToAddress(fee.ContractAddress)
	bz := k.cdc.MustMarshal(&fee)
	store.Set(key.Bytes(), bz)
}

// DeleteFee removes a registered contract
func (k Keeper) DeleteFee(ctx sdk.Context, fee types.DevFeeInfo) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	key := common.HexToAddress(fee.ContractAddress)
	store.Delete(key.Bytes())
}

// IsFeeRegistered - check if registered DevFeeInfo is registered
func (k Keeper) IsFeeRegistered(
	ctx sdk.Context,
	contract common.Address,
) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	return store.Has(contract.Bytes())
}

// GetFeesInverseRaw returns all contracts registered by a deployer as
// types.DevFeeInfosPerDeployer
func (k Keeper) GetFeesInverseRaw(ctx sdk.Context, deployerAddress sdk.AccAddress) (types.DevFeeInfosPerDeployer, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixInverse)
	bz := store.Get(deployerAddress.Bytes())
	if len(bz) == 0 {
		return types.DevFeeInfosPerDeployer{}, false
	}
	var addressList types.DevFeeInfosPerDeployer
	k.cdc.MustUnmarshal(bz, &addressList)
	return addressList, true
}

// GetFeesInverse returns all contracts registered by a deployer as []common.Address
func (k Keeper) GetFeesInverse(ctx sdk.Context, deployerAddress sdk.AccAddress) []common.Address {
	var addresses []common.Address
	addressList, found := k.GetFeesInverseRaw(ctx, deployerAddress)
	if !found {
		return addresses
	}

	for _, addr := range addressList.ContractAddresses {
		addresses = append(addresses, common.HexToAddress(addr))
	}
	return addresses
}

// SetFeeInverse stores a registered contract inverse mapping
func (k Keeper) SetFeeInverse(ctx sdk.Context, deployerAddress sdk.AccAddress, contractAddress common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixInverse)

	store.Set(deployerAddress.Bytes(), contractAddress.Bytes())
}

// DeleteFeeInverse removes a registered contract inverse mapping
func (k Keeper) DeleteFeeInverse(ctx sdk.Context, deployerAddress sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	store.Delete(deployerAddress.Bytes())
}

// HasFeeInverse - check if a reverse mapping deployer => contracts exists
func (k Keeper) HasFeeInverse(
	ctx sdk.Context,
	deployerAddress sdk.AccAddress,
) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixInverse)
	return store.Has(deployerAddress.Bytes())
}
