package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/tharsis/evmos/v3/x/fees/types"
)

// GetAllFees - get all registered DevFeeInfo instances
func (k Keeper) GetAllFees(ctx sdk.Context) []types.DevFeeInfo {
	feeInfos := []types.DevFeeInfo{}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixFee)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		contractAddress := common.BytesToAddress(iterator.Key())
		deployerAddress := sdk.AccAddress(iterator.Value())
		withdrawalAddress, _ := k.GetWithdrawal(ctx, contractAddress)
		feeInfo := types.DevFeeInfo{
			ContractAddress: contractAddress.String(),
			DeployerAddress: deployerAddress.String(),
			WithdrawAddress: withdrawalAddress.String(),
		}
		feeInfos = append(feeInfos, feeInfo)
	}

	return feeInfos
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
		contractAddress := common.BytesToAddress(iterator.Key())
		deployerAddress := sdk.AccAddress(iterator.Value())
		withdrawalAddress, _ := k.GetWithdrawal(ctx, contractAddress)
		feeInfo := types.DevFeeInfo{
			ContractAddress: contractAddress.String(),
			DeployerAddress: deployerAddress.String(),
			WithdrawAddress: withdrawalAddress.String(),
		}

		if handlerFn(feeInfo) {
			break
		}
	}
}

// GetFeeInfo returns DevFeeInfo for a registered contract
func (k Keeper) GetFeeInfo(ctx sdk.Context, contract common.Address) (types.DevFeeInfo, bool) {
	deployerAddress, found := k.GetDeployer(ctx, contract)
	if !found {
		return types.DevFeeInfo{}, false
	}
	withdrawalAddress, _ := k.GetWithdrawal(ctx, contract)
	feeInfo := types.DevFeeInfo{
		ContractAddress: contract.String(),
		DeployerAddress: deployerAddress.String(),
		WithdrawAddress: withdrawalAddress.String(),
	}
	return feeInfo, true
}

// GetDeployer returns the deployer address for a registered contract
func (k Keeper) GetDeployer(ctx sdk.Context, contract common.Address) (sdk.AccAddress, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	bz := store.Get(contract.Bytes())
	if len(bz) == 0 {
		return sdk.AccAddress{}, false
	}
	return sdk.AccAddress(bz), true
}

// GetWithdrawal returns the withdrawal address for a registered contract
func (k Keeper) GetWithdrawal(ctx sdk.Context, contract common.Address) (sdk.AccAddress, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFeeWithdrawal)
	bz := store.Get(contract.Bytes())
	if len(bz) == 0 {
		return k.GetDeployer(ctx, contract)
	}
	return sdk.AccAddress(bz), true
}

// SetFee stores a registered contract
func (k Keeper) SetFee(ctx sdk.Context, contract common.Address, deployer sdk.AccAddress, withdrawal sdk.AccAddress) {
	k.SetDeployer(ctx, contract, deployer)
	if withdrawal != nil {
		k.SetWithdrawal(ctx, contract, withdrawal)
	}
}

// SetDeployer stores the deployer address for a registered contract
func (k Keeper) SetDeployer(ctx sdk.Context, contract common.Address, deployer sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	store.Set(contract.Bytes(), deployer.Bytes())
}

// SetWithdrawal stores the withdrawal address for a registered contract
func (k Keeper) SetWithdrawal(ctx sdk.Context, contract common.Address, withdrawal sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFeeWithdrawal)
	store.Set(contract.Bytes(), withdrawal.Bytes())
}

// DeleteFee removes a registered contract
func (k Keeper) DeleteFee(ctx sdk.Context, contract common.Address) {
	k.DeleteDeployer(ctx, contract)
	k.DeleteWithdrawal(ctx, contract)
}

// DeleteDeployer deletes the deployer address for a registered contract
func (k Keeper) DeleteDeployer(ctx sdk.Context, contract common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	store.Delete(contract.Bytes())
}

// DeleteWithdrawal deletes the withdrawal address for a registered contract
func (k Keeper) DeleteWithdrawal(ctx sdk.Context, contract common.Address) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFeeWithdrawal)
	store.Delete(contract.Bytes())
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
