// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v18/x/evm/types"
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

type Precompiles struct {
	Map       map[common.Address]vm.PrecompiledContract
	Addresses []common.Address
}

// GetPrecompileInstance returns the address and instance of the static or dynamic precompile associated with the given address, or return nil if not found.
func (k *Keeper) GetPrecompileInstance(
	ctx sdktypes.Context,
	address common.Address,
) (*Precompiles, bool, error) {
	params := k.GetParams(ctx)
	// Get the precompile from the static precompiles
	if precompile, found, err := k.GetStaticPrecompileInstance(&params, address); err != nil {
		return nil, false, err
	} else if found {
		addressMap := make(map[common.Address]vm.PrecompiledContract)
		addressMap[address] = precompile
		return &Precompiles{
			Map:       addressMap,
			Addresses: []common.Address{precompile.Address()},
		}, found, nil
	}

	// Get the precompile from the dynamic precompiles
	precompile, found, err := k.GetDynamicPrecompileInstance(ctx, &params, address)
	if err != nil || !found {
		return nil, false, err
	}
	addressMap := make(map[common.Address]vm.PrecompiledContract)
	addressMap[address] = precompile
	return &Precompiles{
		Map:       addressMap,
		Addresses: []common.Address{precompile.Address()},
	}, found, nil
}
