// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	// errorsmod "cosmossdk.io/errors"
	// "github.com/ethereum/go-ethereum/common"
	// "github.com/ethereum/go-ethereum/core/vm"
	// "github.com/evmos/evmos/v18/x/evm/types"
	// "golang.org/x/exp/slices"
	//
	// sdk "github.com/cosmos/cosmos-sdk/types"
)

// // GetDynamicPrecompileInstance returns the dynamic precompile instance for the given address.
// func (k Keeper) GetDynamicPrecompileInstance(
// 	ctx sdk.Context,
// 	params *types.Params,
// 	address common.Address,
// ) (contract vm.PrecompiledContract, found bool, err error) {
// 	if k.IsAvailableDynamicPrecompile(params, address) {
// 		precompile, err := k.erc20Keeper.InstantiateERC20Precompile(ctx, address)
// 		if err != nil {
// 			return nil, false, errorsmod.Wrapf(err, "precompiled contract not initialized: %s", address.String())
// 		}
// 		return precompile, true, nil
// 	}
// 	return nil, false, nil
// }
//
// // IsAvailableDynamicPrecompile returns true if the given precompile address is contained in the
// // EVM keeper's available dynamic precompiles precompiles params.
// func (k Keeper) IsAvailableDynamicPrecompile(params *types.Params, address common.Address) bool {
// 	return slices.Contains(params.WrappedNativeCoinPrecompiles, address.Hex()) ||
// 		slices.Contains(params.ActiveDynamicPrecompiles, address.Hex())
// }
