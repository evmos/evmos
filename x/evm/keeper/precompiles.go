// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v16/x/evm/types"
	"golang.org/x/exp/maps"
)

// GetActivePrecompilesInstances returns a map of both static and dynamic active precompiles
func (k *Keeper) GetActivePrecompilesInstances(
	ctx sdktypes.Context,
	params types.Params,
) ([]common.Address, map[common.Address]vm.PrecompiledContract) {
	staticAddresses, staticPrecompilesMap := k.GetStaticPrecompilesInstances(&params)
	dynamicAddresses, dynamicPrecompileMap := k.GetDynamicPrecompilesInstances(ctx, &params)
	// Append the dynamic precompiles to the active precompiles
	maps.Copy(staticPrecompilesMap, dynamicPrecompileMap)

	// Append the dynamic precompiles to the active precompiles addresses
	staticLen := len(staticAddresses)
	dynamicLen := len(dynamicAddresses)
	totalLen := staticLen + dynamicLen
	addresses := make([]common.Address, totalLen)
	copy(addresses[:staticLen], staticAddresses)
	copy(addresses[staticLen:], dynamicAddresses)

	return addresses, staticPrecompilesMap
}

// GetPrecompileInstance returns the address and instance of the static or dynamic precompile associated with the given address, or return nil if not found.
func (k *Keeper) GetPrecompileInstance(
	ctx sdktypes.Context,
	address common.Address,
) ([]common.Address, map[common.Address]vm.PrecompiledContract, bool) {
	params := k.GetParams(ctx)
	// Get the precompile from the static precompiles
	if precompile, ok := k.GetStaticPrecompileInstance(&params, address); ok {
		addressMap := make(map[common.Address]vm.PrecompiledContract)
		addressMap[address] = precompile
		return []common.Address{precompile.Address()}, addressMap, ok
	}

	// Get the precompile from the dynamic precompiles
	if precompile, ok := k.GetDynamicPrecompileInstance(ctx, &params, address); ok {
		addressMap := make(map[common.Address]vm.PrecompiledContract)
		addressMap[address] = precompile
		return []common.Address{precompile.Address()}, addressMap, ok
	}

	return nil, nil, false
}
