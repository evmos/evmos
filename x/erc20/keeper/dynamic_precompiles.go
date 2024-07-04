// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"
	"slices"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v18/x/erc20/types"
)

// RegisterERC20Extension creates and adds an ERC20 precompile interface for an IBC Coin.
//
// It derives the ERC-20 address from the token denomination and registers the
// EVM extension as an active dynamic precompile.
//
// CONTRACT: This must ONLY be called if there is no existing token pair for the given denom.
func (k Keeper) RegisterERC20Extension(ctx sdk.Context, denom string) (*types.TokenPair, error) {
	pair, err := k.CreateNewTokenPair(ctx, denom, types.OWNER_MODULE)
	if err != nil {
		return nil, err
	}
	// Add to existing EVM extensions
	err = k.EnableDynamicPrecompiles(ctx, pair.GetERC20Contract())
	if err != nil {
		return nil, err
	}
	return &pair, err
}

// EnableDynamicPrecompiles appends the addresses of the given Precompiles to the list
// of active dynamic precompiles.
func (k Keeper) EnableDynamicPrecompiles(ctx sdk.Context, addresses ...common.Address) error {
	// Get the current params and append the new precompiles
	params := k.GetParams(ctx)
	activePrecompiles := params.DynamicPrecompiles

	// Append and sort the new precompiles
	updatedPrecompiles, err := appendPrecompiles(activePrecompiles, addresses...)
	if err != nil {
		return err
	}

	// Update params
	params.DynamicPrecompiles = updatedPrecompiles
	k.Logger(ctx).Info("Added new precompiles", "addresses", addresses)
	return k.SetParams(ctx, params)
}

func appendPrecompiles(existingPrecompiles []string, addresses ...common.Address) ([]string, error) {
	// check for duplicates
	hexAddresses := make([]string, len(addresses))
	for i := range addresses {
		addrHex := addresses[i].Hex()
		if slices.Contains(existingPrecompiles, addrHex) {
			return nil, fmt.Errorf("precompile already registered: %s", addrHex)
		}
		hexAddresses[i] = addrHex
	}

	exstingLength := len(existingPrecompiles)
	updatedPrecompiles := make([]string, exstingLength+len(hexAddresses))
	copy(updatedPrecompiles, existingPrecompiles)
	copy(updatedPrecompiles[exstingLength:], hexAddresses)

	sortPrecompiles(updatedPrecompiles)
	return updatedPrecompiles, nil
}

func sortPrecompiles(precompiles []string) {
	sort.Slice(precompiles, func(i, j int) bool {
		return precompiles[i] < precompiles[j]
	})
}
