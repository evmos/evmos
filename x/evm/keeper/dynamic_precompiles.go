// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AddDynamicPrecompiles adds the given precompiles to the list of active precompiles
func (k *Keeper) AddDynamicPrecompiles(ctx sdk.Context, precompiles ...vm.PrecompiledContract) error {
	addresses := make([]common.Address, len(precompiles))
	for i, precompile := range precompiles {
		address := precompile.Address()
		addresses[i] = address
	}

	return k.EnableDynamicPrecompiles(ctx, addresses...)
}

// GetDynamicPrecompileInstance returns a map of active precompiles
func (k Keeper) GetDynamicPrecompileInstance(
	ctx sdk.Context,
	activePrecompiles ...common.Address,
) map[common.Address]vm.PrecompiledContract {
	activePrecompileMap := make(map[common.Address]vm.PrecompiledContract)
	for _, address := range activePrecompiles {
		precompile, err := k.erc20Keeper.InstantiateERC20Precompile(ctx, address)
		if err != nil {
			panic(errorsmod.Wrapf(err, "precompiled contract not initialized: %s", address))
		}

		activePrecompileMap[address] = precompile
	}
	return activePrecompileMap
}
