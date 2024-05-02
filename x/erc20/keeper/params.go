// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v18/x/erc20/types"
	"slices"
)

var isTrue = []byte("0x01")

// GetParams returns the total set of erc20 parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixParams)
	if len(bz) == 0 {
		panic("ERC20 params not found")
	}
	k.cdc.MustUnmarshal(bz, &params)
	return
	// enableErc20 := k.IsERC20Enabled(ctx)
	// enableEvmHook := k.GetEnableEVMHook(ctx)
	//
	// return types.NewParams(enableErc20, enableEvmHook)
}

// TODO - DO NOT LET ME MERGE THIS
// SetParams sets the erc20 parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	// and keep params equal between different executions
	slices.Sort(params.DynamicPrecompiles)
	slices.Sort(params.NativePrecompiles)

	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(types.KeyPrefixParams, bz)
	return nil

	// k.setERC20Enabled(ctx, params.EnableErc20)
	// k.setEnableEVMHook(ctx, params.EnableEVMHook)
	//
	// return nil
}

// IsERC20Enabled returns true if the module logic is enabled
func (k Keeper) IsERC20Enabled(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.ParamStoreKeyEnableErc20)
}

// GetEnableEVMHook returns true if the EVM hooks are enabled
func (k Keeper) GetEnableEVMHook(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.ParamStoreKeyEnableEVMHook)
}

// setERC20Enabled sets the EnableERC20 param in the store
func (k Keeper) setERC20Enabled(ctx sdk.Context, enable bool) {
	store := ctx.KVStore(k.storeKey)
	if enable {
		store.Set(types.ParamStoreKeyEnableErc20, isTrue)
		return
	}
	store.Delete(types.ParamStoreKeyEnableErc20)
}

// setEnableEVMHook sets the EnableEVMHook param in the store
func (k Keeper) setEnableEVMHook(ctx sdk.Context, enable bool) {
	store := ctx.KVStore(k.storeKey)
	if enable {
		store.Set(types.ParamStoreKeyEnableEVMHook, isTrue)
		return
	}
	store.Delete(types.ParamStoreKeyEnableEVMHook)
}
