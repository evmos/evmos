// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v16/x/evm/types"
	"golang.org/x/exp/slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AddDynamicPrecompiles adds the given precompiles to the list of active precompiles
func (k *Keeper) AddDynamicPrecompiles(ctx sdk.Context, precompiles ...vm.PrecompiledContract) error {
	addresses := make([]string, len(precompiles))
	for i, precompile := range precompiles {
		address := precompile.Address()
		addresses[i] = address.String()
	}

	return k.EnableDynamicPrecompiles(ctx, addresses...)
}

// GetDynamicPrecompilesInstances returns the addresses and instances of the active dynamic precompiles
func (k Keeper) GetDynamicPrecompilesInstances(
	ctx sdk.Context,
	params *types.Params,
) ([]common.Address, map[common.Address]vm.PrecompiledContract) {
	activePrecompileMap := make(map[common.Address]vm.PrecompiledContract)
	addresses := make([]common.Address, len(params.ActiveDynamicPrecompiles))

	for i, address := range params.ActiveDynamicPrecompiles {
		hexAddress := common.HexToAddress(address)

		precompile, err := k.erc20Keeper.InstantiateERC20Precompile(ctx, hexAddress)
		if err != nil {
			panic(errorsmod.Wrapf(err, "precompiled contract not initialized: %s", address))
		}

		activePrecompileMap[hexAddress] = precompile
		addresses[i] = hexAddress
	}
	return addresses, activePrecompileMap
}

// IsAvailableDynamicPrecompile returns true if the given precompile address is contained in the
// EVM keeper's available dynamic precompiles precompiles params.
func (k Keeper) IsAvailableDynamicPrecompile(ctx sdk.Context, address string) bool {
	return slices.Contains(k.GetParams(ctx).ActiveDynamicPrecompiles, address)
}
