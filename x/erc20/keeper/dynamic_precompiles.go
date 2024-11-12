// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/evmos/v20/utils"
	"github.com/evmos/evmos/v20/x/erc20/types"
	"github.com/evmos/evmos/v20/x/evm/statedb"
)

// RegisterERC20Extension creates and adds an ERC20 precompile interface for an IBC Coin.
//
// It derives the ERC-20 address from the token denomination and registers the
// EVM extension as an active dynamic precompile.
//
// CONTRACT: This must ONLY be called if there is no existing token pair for the given denom.
func (k Keeper) RegisterERC20Extension(ctx sdk.Context, denom string) (*types.TokenPair, error) {
	pair, err := k.CreateNewTokenPair(ctx, denom)
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

// RegisterERC20CodeHash sets the codehash for the erc20 precompile account
// if the bytecode for the erc20 codehash does not exists, it stores it.
func (k Keeper) RegisterERC20CodeHash(ctx sdk.Context, erc20Addr common.Address) error {
	var (
		// bytecode and codeHash is the same for all IBC coins
		// cause they're all using the same contract
		bytecode = common.FromHex(types.Erc20Bytecode)
		codeHash = crypto.Keccak256(bytecode)
	)
	// check if code was already stored
	code := k.evmKeeper.GetCode(ctx, common.Hash(codeHash))
	if len(code) == 0 {
		k.evmKeeper.SetCode(ctx, codeHash, bytecode)
	}

	var (
		nonce   uint64
		balance = common.Big0
	)
	// keep balance and nonce if account exists
	if acc := k.evmKeeper.GetAccount(ctx, erc20Addr); acc != nil {
		nonce = acc.Nonce
		balance = acc.Balance
	}

	return k.evmKeeper.SetAccount(ctx, erc20Addr, statedb.Account{
		CodeHash: codeHash,
		Nonce:    nonce,
		Balance:  balance,
	})
}

// UnRegisterERC20CodeHash sets the codehash for the account to an empty one
func (k Keeper) UnRegisterERC20CodeHash(ctx sdk.Context, erc20Addr common.Address) error {
	emptyCodeHash := crypto.Keccak256(nil)

	var (
		nonce   uint64
		balance = common.Big0
	)
	// keep balance and nonce if account exists
	if acc := k.evmKeeper.GetAccount(ctx, erc20Addr); acc != nil {
		nonce = acc.Nonce
		balance = acc.Balance
	}

	return k.evmKeeper.SetAccount(ctx, erc20Addr, statedb.Account{
		CodeHash: emptyCodeHash,
		Nonce:    nonce,
		Balance:  balance,
	})
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

// appendPrecompiles append addresses to the existingPrecompiles and sort the resulting slice.
// The function returns an error is the two sets are overlapping.
func appendPrecompiles(existingPrecompiles []string, addresses ...common.Address) ([]string, error) {
	// check for duplicates
	hexAddresses := make([]string, len(addresses))
	for i := range addresses {
		addrHex := addresses[i].Hex()
		if slices.Contains(existingPrecompiles, addrHex) {
			return nil, fmt.Errorf("attempted to register a duplicate precompile address: %s", addrHex)
		}
		hexAddresses[i] = addrHex
	}

	existingLength := len(existingPrecompiles)
	updatedPrecompiles := make([]string, existingLength+len(hexAddresses))
	copy(updatedPrecompiles, existingPrecompiles)
	copy(updatedPrecompiles[existingLength:], hexAddresses)

	utils.SortSlice(updatedPrecompiles)
	return updatedPrecompiles, nil
}
