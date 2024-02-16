// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v16/precompiles/erc20"
	"github.com/evmos/evmos/v16/x/erc20/types"
)

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

// RegisterERC20Extension Creates and adds an ERC20 precompile interface for an IBC Coin.
// It truncates the denom address to 20 bytes and registers the precompile if it is not already registered
func (k Keeper) RegisterERC20Extension(ctx sdk.Context, denom string, contractAddr common.Address) error {
	pair, err := k.CreateNewTokenPair(ctx, denom, types.OWNER_MODULE)
	if err != nil {
		return err
	}

	// Add to existing EVM extensions
	return k.evmKeeper.EnableDynamicPrecompiles(ctx, pair.GetERC20Contract())
}
