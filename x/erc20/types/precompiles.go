// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"fmt"
	"slices"

	"github.com/evmos/evmos/v18/types"
)

const (
	// WEVMOSContractMainnet is the WEVMOS contract address for mainnet
	WEVMOSContractMainnet = "0xD4949664cD82660AaE99bEdc034a0deA8A0bd517"
	// WEVMOSContractTestnet is the WEVMOS contract address for testnet
	WEVMOSContractTestnet = "0xcc491f589b45d4a3c679016195b3fb87d7848210"
)

// Precompile store key
var (
	PrecompileStoreKeyDynamic = []byte("Dynamic")
	PrecompileStoreKeyNative  = []byte("Native")
)

// Default Values
var (
	// DefaultNativePrecompiles defines the default precompiles for the wrapped native coin
	// NOTE: If you modify this, make sure you modify it on the local_node genesis script as well
	DefaultNativePrecompiles = []string{WEVMOSContractMainnet}
	// DefaultDynamicPrecompiles defines the default active dynamic precompiles
	DefaultDynamicPrecompiles []string
)

// NewParams creates a new Params object
func NewPrecompiles(
	nativePrecompiles []string,
	dynamicPrecompiles []string,
) Precompiles {
	slices.Sort(nativePrecompiles)
	slices.Sort(dynamicPrecompiles)
	return Precompiles{
		Native:  nativePrecompiles,
		Dynamic: dynamicPrecompiles,
	}
}

func DefaultPrecompiles() Precompiles {
	return Precompiles{
		Native:  DefaultNativePrecompiles,
		Dynamic: DefaultDynamicPrecompiles,
	}
}

func (p Precompiles) Validate() error {
	if err := ValidatePrecompiles(p.Native); err != nil {
		return err
	}

	if err := ValidatePrecompiles(p.Dynamic); err != nil {
		return err
	}

    combined := append(p.Dynamic, p.Native...)
	return ValidatePrecompilesUniqueness(combined)
}

// ValidatePrecompiles checks if the precompile addresses are valid and unique.
func ValidatePrecompiles(i interface{}) error {
	precompiles, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid precompile slice type: %T", i)
	}

	for _, precompile := range precompiles {
		if err := types.ValidateAddress(precompile); err != nil {
			return fmt.Errorf("invalid precompile %s", precompile)
		}
	}

	// NOTE: Check that the precompiles are sorted. This is required for the
	// precompiles to be found correctly when using the IsActivePrecompile method,
	// because of the use of sort.Find.
	if !slices.IsSorted(precompiles) {
		return fmt.Errorf("precompiles need to be sorted: %s", precompiles)
	}
	return nil
}

func ValidatePrecompilesUniqueness(i interface{}) error {
	precompiles, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid precompile slice type: %T", i)
	}

	seenPrecompiles := make(map[string]struct{})
	for _, precompile := range precompiles {
		if _, ok := seenPrecompiles[precompile]; ok {
			return fmt.Errorf("duplicate precompile %s", precompile)
		}

		seenPrecompiles[precompile] = struct{}{}
	}
	return nil
}
