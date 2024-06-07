// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"
	"slices"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v18/precompiles/erc20"
	"github.com/evmos/evmos/v18/x/erc20/types"
)

// GetDynamicPrecompileInstance returns the dynamic precompile instance for the given address.
func (k Keeper) GetERC20PrecompileInstance(
	ctx sdk.Context,
	address common.Address,
) (contract vm.PrecompiledContract, found bool, err error) {
	precompiles := k.GetPrecompiles(ctx)

	if k.IsAvailableERC20Precompile(&precompiles, address) {
		precompile, err := k.InstantiateERC20Precompile(ctx, address)
		if err != nil {
			return nil, false, errorsmod.Wrapf(err, "precompiled contract not initialized: %s", address.String())
		}
		return precompile, true, nil
	}
	return nil, false, nil
}

// InstantiateERC20Precompile returns an ERC20 precompile instance for the given contract address
func (k Keeper) InstantiateERC20Precompile(ctx sdk.Context, contractAddr common.Address) (vm.PrecompiledContract, error) {
	address := contractAddr.String()
	// check if the precompile is an ERC20 contract
	id := k.GetTokenPairID(ctx, address)
	if len(id) == 0 {
		return nil, fmt.Errorf("precompile id not found: %s", address)
	}
	pair, ok := k.GetTokenPair(ctx, id)
	if !ok {
		return nil, fmt.Errorf("token pair not found: %s", address)
	}
	return erc20.NewPrecompile(pair, k.bankKeeper, k.authzKeeper, *k.transferKeeper)
}

// IsAvailableDynamicPrecompile returns true if the given precompile address is contained in the
// EVM keeper's available dynamic precompiles precompiles params.
func (k Keeper) IsAvailableERC20Precompile(precompiles *types.Precompiles, address common.Address) bool {
	return slices.Contains(precompiles.Native, address.Hex()) ||
		slices.Contains(precompiles.Dynamic, address.Hex())
}
func (k Keeper) GetPrecompiles(ctx sdk.Context) types.Precompiles {
	dynamicPrecompiles := k.getDynamicPrecompiles(ctx)
	nativePrecompiles := k.getNativePrecompiles(ctx)
	return types.NewPrecompiles(nativePrecompiles, dynamicPrecompiles)
}

// SetPrecompiles sets the erc20 precompiles.
func (k Keeper) SetPrecompiles(ctx sdk.Context, precompiles types.Precompiles) error {
	// sort and keep params equal between different executions
	slices.Sort(precompiles.Dynamic)
	slices.Sort(precompiles.Native)

	if err := precompiles.Validate(); err != nil {
		return err
	}

	k.setDynamicPrecompiles(ctx, precompiles.Dynamic)
	k.setNativePrecompiles(ctx, precompiles.Native)
	return nil
}

func (k Keeper) setNativePrecompiles(ctx sdk.Context, nativePrecompiles []string) {
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 0, addressLength*len(nativePrecompiles))
	for _, str := range nativePrecompiles {
		bz = append(bz, []byte(str)...)
	}
	store.Set(types.PrecompileStoreKeyNative, bz)
}

func (k Keeper) getNativePrecompiles(ctx sdk.Context) (nativePrecompiles []string) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.PrecompileStoreKeyNative)
	for i := 0; i < len(bz); i += addressLength {
		nativePrecompiles = append(nativePrecompiles, string(bz[i:i+addressLength]))
	}
	return nativePrecompiles
}

func (k Keeper) setDynamicPrecompiles(ctx sdk.Context, dynamicPrecompiles []string) {
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 0, addressLength*len(dynamicPrecompiles))
	for _, str := range dynamicPrecompiles {
		bz = append(bz, []byte(str)...)
	}
	store.Set(types.PrecompileStoreKeyDynamic, bz)
}

func (k Keeper) getDynamicPrecompiles(ctx sdk.Context) (precompiles []string) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.PrecompileStoreKeyDynamic)
	for i := 0; i < len(bz); i += addressLength {
		precompiles = append(precompiles, string(bz[i:i+addressLength]))
	}
	return precompiles
}
