// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"fmt"

	"github.com/evmos/evmos/v18/utils"
)

var DefaultTokenPairs = []TokenPair{
	{
		Erc20Address:  WEVMOSContractMainnet,
		Denom:         utils.BaseDenom,
		Enabled:       true,
		ContractOwner: OWNER_MODULE,
	},
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, pairs []TokenPair) GenesisState {
	return GenesisState{
		Params:     params,
		TokenPairs: pairs,
	}
}

// DefaultGenesisState sets default evm genesis state with empty accounts and
// default params and chain config values.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:     DefaultParams(),
		TokenPairs: DefaultTokenPairs,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
// TODO: Validate that the precompiles have a corresponding token pair
func (gs GenesisState) Validate() error {
	seenErc20 := make(map[string]bool)
	seenDenom := make(map[string]bool)

	for _, b := range gs.TokenPairs {
		if seenErc20[b.Erc20Address] {
			return fmt.Errorf("token ERC20 contract duplicated on genesis '%s'", b.Erc20Address)
		}
		if seenDenom[b.Denom] {
			return fmt.Errorf("coin denomination duplicated on genesis: '%s'", b.Denom)
		}

		if err := b.Validate(); err != nil {
			return err
		}

		seenErc20[b.Erc20Address] = true
		seenDenom[b.Denom] = true
	}

	// Check if params are valid
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params on genesis: %w", err)
	}

	// Check if active precompiles have a corresponding token pair
	if err := validatePrecompiles(gs.TokenPairs, gs.Params.DynamicPrecompiles); err != nil {
		return fmt.Errorf("invalid dynamic precompiles on genesis: %w", err)
	}

	if err := validatePrecompiles(gs.TokenPairs, gs.Params.NativePrecompiles); err != nil {
		return fmt.Errorf("invalid native precompiles on genesis: %w", err)
	}
	return nil
}

// validatePrecompiles checks if every precompile has a corresponding enabled token pair
func validatePrecompiles(tokenPairs []TokenPair, precompiles []string) error {
	for _, precompile := range precompiles {
		if !hasActiveTokenPair(tokenPairs, precompile) {
			return fmt.Errorf("precompile address '%s' not found in token pairs", precompile)
		}
	}
	return nil
}

func hasActiveTokenPair(pairs []TokenPair, address string) bool {
	for _, p := range pairs {
		if p.Erc20Address == address && p.Enabled {
			return true
		}
	}
	return false
}
