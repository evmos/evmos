package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/tharsis/evmos/v5/x/fees/types"
)

// GetAllFees - get all registered DevFeeInfo instances
func (k Keeper) GetAllFees(ctx sdk.Context) []types.DevFeeInfo {
	feeInfos := []types.DevFeeInfo{}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixFee)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		feeInfo := k.BuildFeeInfo(
			ctx,
			common.BytesToAddress(iterator.Key()),
			sdk.AccAddress(iterator.Value()),
		)
		feeInfos = append(feeInfos, feeInfo)
	}

	return feeInfos
}

// IterateFees iterates over all registered contracts and performs a
// callback with the corresponding DevFeeInfo.
func (k Keeper) IterateFees(
	ctx sdk.Context,
	handlerFn func(fee types.DevFeeInfo) (stop bool),
) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixFee)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		feeInfo := k.BuildFeeInfo(
			ctx,
			common.BytesToAddress(iterator.Key()),
			sdk.AccAddress(iterator.Value()),
		)
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
	feeInfo := k.BuildFeeInfo(ctx, contract, deployerAddress)
	return feeInfo, true
}

// BuildFeeInfo returns DevFeeInfo given the contract and deployer addresses
func (k Keeper) BuildFeeInfo(ctx sdk.Context, contract common.Address, deployerAddress sdk.AccAddress) types.DevFeeInfo {
	withdrawalAddress, hasWithdrawAddr := k.GetWithdrawal(ctx, contract)
	feeInfo := types.DevFeeInfo{
		ContractAddress: contract.String(),
		DeployerAddress: deployerAddress.String(),
	}
	if hasWithdrawAddr {
		feeInfo.WithdrawAddress = withdrawalAddress.String()
	}
	return feeInfo
}

// GetDeployer returns the deployer address for a registered contract
func (k Keeper) GetDeployer(ctx sdk.Context, contract common.Address) (sdk.AccAddress, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	bz := store.Get(contract.Bytes())
	if len(bz) == 0 {
		return nil, false
	}
	return sdk.AccAddress(bz), true
}

// GetWithdrawal returns the withdrawal address for a registered contract
func (k Keeper) GetWithdrawal(ctx sdk.Context, contract common.Address) (sdk.AccAddress, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFeeWithdrawal)
	bz := store.Get(contract.Bytes())
	if len(bz) == 0 {
		return nil, false
	}
	return sdk.AccAddress(bz), true
}

// SetFee stores the developer fee information for a registered contract
func (k Keeper) SetFee(ctx sdk.Context, contract common.Address, deployer sdk.AccAddress, withdrawal sdk.AccAddress) {
	k.SetDeployer(ctx, contract, deployer)
	if len(withdrawal) > 0 && withdrawal.String() != deployer.String() {
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

// IsFeeRegistered checks if a contract was registered for receiving fees
func (k Keeper) IsFeeRegistered(
	ctx sdk.Context,
	contract common.Address,
) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
	return store.Has(contract.Bytes())
}

// GetFeesInverse returns all contracts registered by a deployer as []common.Address
func (k Keeper) GetFeesInverse(ctx sdk.Context, deployerAddress sdk.AccAddress) []common.Address {
	feeInfoKeys := []common.Address{}
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(
		store,
		types.GetKeyPrefixInverseDeployer(deployerAddress),
	)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		feeInfoKeys = append(feeInfoKeys, common.BytesToAddress(iterator.Key()))
	}

	return feeInfoKeys
}

// SetFeeInverse stores a registered contract inverse mapping
func (k Keeper) SetFeeInverse(ctx sdk.Context, deployerAddress sdk.AccAddress, contractAddress common.Address) {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.GetKeyPrefixInverseDeployer(deployerAddress),
	)
	store.Set(contractAddress.Bytes(), []byte("1"))
}

// DeleteFeeInverse removes a registered contract from a deployer's KVStore of
// registered contracts
func (k Keeper) DeleteFeeInverse(ctx sdk.Context, deployerAddress sdk.AccAddress, contractAddress common.Address) {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.GetKeyPrefixInverseDeployer(deployerAddress),
	)
	store.Delete(contractAddress.Bytes())
}

// IsFeeInverseRegistered checks if a contract exists in a deployer's KVStore of
// registered contracts
func (k Keeper) IsFeeInverseRegistered(
	ctx sdk.Context,
	deployerAddress sdk.AccAddress,
	contractAddress common.Address,
) bool {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.GetKeyPrefixInverseDeployer(deployerAddress),
	)
	return store.Has(contractAddress.Bytes())
}
