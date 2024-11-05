// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/x/erc20/types"
)

var isTrue = []byte("0x01")

const addressLength = 42

// GetParams returns the total set of erc20 parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	enableErc20 := k.IsERC20Enabled(ctx)
	dynamicPrecompiles := k.getDynamicPrecompiles(ctx)
	nativePrecompiles := k.getNativePrecompiles(ctx)
	return types.NewParams(enableErc20, nativePrecompiles, dynamicPrecompiles)
}

// UpdateCodeHash takes in the updated parameters and
// compares the new set of native and dynamic precompiles to the current
// parameter set.
//
// If there is a diff, the ERC-20 code hash for all precompiles that are removed from the list
// will be removed from the store. Meanwhile, for all newly added precompiles the code hash will be
// registered.
func (k Keeper) UpdateCodeHash(ctx sdk.Context, newParams types.Params) error {
	oldNativePrecompiles := k.getNativePrecompiles(ctx)
	oldDynamicPrecompiles := k.getDynamicPrecompiles(ctx)

	if err := k.RegisterOrUnregisterERC20CodeHashes(ctx, oldDynamicPrecompiles, newParams.DynamicPrecompiles); err != nil {
		return err
	}

	return k.RegisterOrUnregisterERC20CodeHashes(ctx, oldNativePrecompiles, newParams.NativePrecompiles)
}

// RegisterOrUnregisterERC20CodeHashes takes two arrays of precompiles as its argument:
//   - previously registered precompiles
//   - new set of precompiles to be registered
//
// It then compares the two arrays and registers the code hash for all precompiles that are newly added
// and unregisters the code hash for all precompiles that are removed from the list.
func (k Keeper) RegisterOrUnregisterERC20CodeHashes(ctx sdk.Context, oldPrecompiles, newPrecompiles []string) error {
	for _, precompile := range oldPrecompiles {
		if slices.Contains(newPrecompiles, precompile) {
			continue
		}

		if err := k.UnRegisterERC20CodeHash(ctx, common.HexToAddress(precompile)); err != nil {
			return err
		}
	}

	for _, precompile := range newPrecompiles {
		if slices.Contains(oldPrecompiles, precompile) {
			continue
		}

		if err := k.RegisterERC20CodeHash(ctx, common.HexToAddress(precompile)); err != nil {
			return err
		}
	}

	return nil
}

// SetParams sets the erc20 parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, newParams types.Params) error {
	// sort to keep params equal between different executions
	slices.Sort(newParams.DynamicPrecompiles)
	slices.Sort(newParams.NativePrecompiles)

	if err := newParams.Validate(); err != nil {
		return err
	}

	if err := k.UpdateCodeHash(ctx, newParams); err != nil {
		return err
	}

	k.setERC20Enabled(ctx, newParams.EnableErc20)
	k.setDynamicPrecompiles(ctx, newParams.DynamicPrecompiles)
	k.setNativePrecompiles(ctx, newParams.NativePrecompiles)
	return nil
}

// IsERC20Enabled returns true if the module logic is enabled
func (k Keeper) IsERC20Enabled(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.ParamStoreKeyEnableErc20)
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

// setDynamicPrecompiles sets the DynamicPrecompiles param in the store
func (k Keeper) setDynamicPrecompiles(ctx sdk.Context, dynamicPrecompiles []string) {
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 0, addressLength*len(dynamicPrecompiles))
	for _, str := range dynamicPrecompiles {
		bz = append(bz, []byte(str)...)
	}
	store.Set(types.ParamStoreKeyDynamicPrecompiles, bz)
}

// getDynamicPrecompiles returns the DynamicPrecompiles param from the store
func (k Keeper) getDynamicPrecompiles(ctx sdk.Context) (dynamicPrecompiles []string) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyDynamicPrecompiles)

	for i := 0; i < len(bz); i += addressLength {
		dynamicPrecompiles = append(dynamicPrecompiles, string(bz[i:i+addressLength]))
	}
	return dynamicPrecompiles
}

// setNativePrecompiles sets the NativePrecompiles param in the store
func (k Keeper) setNativePrecompiles(ctx sdk.Context, nativePrecompiles []string) {
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 0, addressLength*len(nativePrecompiles))
	for _, str := range nativePrecompiles {
		bz = append(bz, []byte(str)...)
	}
	store.Set(types.ParamStoreKeyNativePrecompiles, bz)
}

// getNativePrecompiles returns the NativePrecompiles param from the store
func (k Keeper) getNativePrecompiles(ctx sdk.Context) (nativePrecompiles []string) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyNativePrecompiles)
	for i := 0; i < len(bz); i += addressLength {
		nativePrecompiles = append(nativePrecompiles, string(bz[i:i+addressLength]))
	}
	return nativePrecompiles
}
