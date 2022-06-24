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
	// TODO move this check out of `SetFee`?
	// prevent storing the same address for deployer and withdrawer
	if fee.DeployerAddress == fee.WithdrawAddress {
		fee.WithdrawAddress = ""
	}

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

// deleteDeployerMap deletes a fee contract by deployer mapping
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

// // GetDeployer returns the deployer address for a registered contract
// func (k Keeper) GetDeployer(ctx sdk.Context, contract common.Address) (sdk.AccAddress, bool) {
// 	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
// 	bz := store.Get(contract.Bytes())
// 	if len(bz) == 0 {
// 		return nil, false
// 	}
// 	return sdk.AccAddress(bz), true
// }

// // SetDeployer stores the deployer address for a registered contract
// func (k Keeper) SetDeployer(ctx sdk.Context, contract common.Address, deployer sdk.AccAddress) {
// 	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
// 	store.Set(contract.Bytes(), deployer.Bytes())
// }

// // DeleteDeployer deletes the deployer address for a registered contract
// func (k Keeper) DeleteDeployer(ctx sdk.Context, contract common.Address) {
// 	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFee)
// 	store.Delete(contract.Bytes())
// }

// // GetWithdrawal returns the withdrawal address for a registered contract
// func (k Keeper) GetWithdrawal(ctx sdk.Context, contract common.Address) (sdk.AccAddress, bool) {
// 	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFeeWithdrawal)
// 	bz := store.Get(contract.Bytes())
// 	if len(bz) == 0 {
// 		return nil, false
// 	}
// 	return sdk.AccAddress(bz), true
// }

// // SetWithdrawal stores the withdrawal address for a registered contract
// func (k Keeper) SetWithdrawal(ctx sdk.Context, contract common.Address, withdrawal sdk.AccAddress) {
// 	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFeeWithdrawal)
// 	store.Set(contract.Bytes(), withdrawal.Bytes())
// }

// // DeleteWithdrawal deletes the withdrawal address for a registered contract
// func (k Keeper) DeleteWithdrawal(ctx sdk.Context, contract common.Address) {
// 	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixFeeWithdrawal)
// 	store.Delete(contract.Bytes())
// }

// TODO Queries for `GetDeployerFees` and `GetWithdrawFees`

// TODO GetDeployerFees returns all contracts registered by a deployer as []common.Address
func (k Keeper) GetDeployerFees(ctx sdk.Context, deployerAddress sdk.AccAddress) []common.Address {
	fees := []common.Address{}
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(
		store,
		types.GetKeyPrefixDeployerFees(deployerAddress),
	)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		fees = append(fees, common.BytesToAddress(iterator.Key()))
	}

	return fees
}

// // SetDeployerFees stores a registered contract inverse mapping
// func (k Keeper) SetDeployerFees(ctx sdk.Context, deployerAddress sdk.AccAddress, contractAddress common.Address) {
// 	store := prefix.NewStore(
// 		ctx.KVStore(k.storeKey),
// 		types.GetKeyPrefixDeployerFees(deployerAddress),
// 	)
// 	store.Set(contractAddress.Bytes(), []byte("1"))
// }

// // DeleteDeployerFees removes a registered contract from a deployer's KVStore of
// // registered contracts
// func (k Keeper) DeleteDeployerFees(ctx sdk.Context, deployerAddress sdk.AccAddress, contractAddress common.Address) {
// 	store := prefix.NewStore(
// 		ctx.KVStore(k.storeKey),
// 		types.GetKeyPrefixDeployerFees(deployerAddress),
// 	)
// 	store.Delete(contractAddress.Bytes())
// }

// // IsDeployerFeesRegistered checks if a contract exists in a deployer's KVStore of
// // registered contracts
// func (k Keeper) IsDeployerFeesRegistered(
// 	ctx sdk.Context,
// 	deployerAddress sdk.AccAddress,
// 	contractAddress common.Address,a
// ) bool {
// 	store := prefix.NewStore(
// 		ctx.KVStore(k.storeKey),
// 		types.GetKeyPrefixDeployerFees(deployerAddress),
// 	)
// 	return store.Has(contractAddress.Bytes())
// }
