// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"
	"slices"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/precompiles/erc20"
	"github.com/evmos/evmos/v19/x/erc20/types"
	"github.com/evmos/evmos/v19/x/evm/core/vm"
)

// GetERC20PrecompileInstance returns the precompile instance for the given address.
func (k Keeper) GetERC20PrecompileInstance(
	ctx sdk.Context,
	address common.Address,
) (contract vm.PrecompiledContract, found bool, err error) {
	params := k.GetParams(ctx)

	if k.IsAvailableERC20Precompile(&params, address) {
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
func (k Keeper) IsAvailableERC20Precompile(params *types.Params, address common.Address) bool {
	return slices.Contains(params.NativePrecompiles, address.Hex()) ||
		slices.Contains(params.DynamicPrecompiles, address.Hex())
}
