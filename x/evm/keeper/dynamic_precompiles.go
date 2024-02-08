// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package keeper

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k *Keeper) AddDynamicPrecompiles(ctx sdk.Context, precompiles ...vm.PrecompiledContract) error {
	addresses := make([]common.Address, len(precompiles))
	for i, precompile := range precompiles {
		address := precompile.Address()
		addresses[i] = address
	}

	err := k.EnableDynamicPrecompiles(ctx, addresses...)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) GetDynamicPrecompileInstance(
	activePrecompiles ...common.Address,
) map[common.Address]vm.PrecompiledContract {
	activePrecompileMap := make(map[common.Address]vm.PrecompiledContract)
	for _, address := range activePrecompiles {
		precompile, ok := k.precompiles[address]
		if !ok {
			panic(fmt.Sprintf("precompiled contract not initialized: %s", address))
		}

		activePrecompileMap[address] = precompile
	}
	return activePrecompileMap
}
