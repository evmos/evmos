package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v5/x/fees/types"
)

// GetFees - get all registered Fee instances
func (k Keeper) GetFees(ctx sdk.Context) []types.Fee {
	fees := []types.Fee{}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixFee)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		fee := k.BuildFeeInfo(
			ctx,
			common.BytesToAddress(iterator.Key()),
			sdk.AccAddress(iterator.Value()),
		)
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
		fee := k.BuildFeeInfo(
			ctx,
			common.BytesToAddress(iterator.Key()),
			sdk.AccAddress(iterator.Value()),
		)
		if handlerFn(fee) {
			break
		}
	}
}

// GetFee returns Fee for a registered contract
func (k Keeper) GetFee(ctx sdk.Context, contract common.Address) (types.Fee, bool) {
	deployerAddress, found := k.GetDeployer(ctx, contract)
	if !found {
		return types.Fee{}, false
	}
	fee := k.BuildFeeInfo(ctx, contract, deployerAddress)
	return fee, true
}

// BuildFeeInfo returns Fee given the contract and deployer addresses
func (k Keeper) BuildFeeInfo(ctx sdk.Context, contract common.Address, deployerAddress sdk.AccAddress) types.Fee {
	withdrawalAddress, hasWithdrawAddr := k.GetWithdrawal(ctx, contract)
	fee := types.Fee{
		ContractAddress: contract.String(),
		DeployerAddress: deployerAddress.String(),
	}
	if hasWithdrawAddr {
		fee.WithdrawAddress = withdrawalAddress.String()
	}
	return fee
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

// GetDeployerFees returns all contracts registered by a deployer as []common.Address
func (k Keeper) GetDeployerFees(ctx sdk.Context, deployerAddress sdk.AccAddress) []common.Address {
	feeKeys := []common.Address{}
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(
		store,
		types.GetKeyPrefixDeployerFees(deployerAddress),
	)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		feeKeys = append(feeKeys, common.BytesToAddress(iterator.Key()))
	}

	return feeKeys
}

// SetDeployerFees stores a registered contract inverse mapping
func (k Keeper) SetDeployerFees(ctx sdk.Context, deployerAddress sdk.AccAddress, contractAddress common.Address) {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.GetKeyPrefixDeployerFees(deployerAddress),
	)
	store.Set(contractAddress.Bytes(), []byte("1"))
}

// DeleteDeployerFees removes a registered contract from a deployer's KVStore of
// registered contracts
func (k Keeper) DeleteDeployerFees(ctx sdk.Context, deployerAddress sdk.AccAddress, contractAddress common.Address) {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.GetKeyPrefixDeployerFees(deployerAddress),
	)
	store.Delete(contractAddress.Bytes())
}

// IsDeployerFeesRegistered checks if a contract exists in a deployer's KVStore of
// registered contracts
func (k Keeper) IsDeployerFeesRegistered(
	ctx sdk.Context,
	deployerAddress sdk.AccAddress,
	contractAddress common.Address,
) bool {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.GetKeyPrefixDeployerFees(deployerAddress),
	)
	return store.Has(contractAddress.Bytes())
}
