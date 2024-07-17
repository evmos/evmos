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

// Parameter store key
var (
	ParamStoreKeyEnableErc20        = []byte("EnableErc20")
	ParamStoreKeyDynamicPrecompiles = []byte("DynamicPrecompiles")
	ParamStoreKeyNativePrecompiles  = []byte("NativePrecompiles")
	// DefaultNativePrecompiles defines the default precompiles for the wrapped native coin
	// NOTE: If you modify this, make sure you modify it on the local_node genesis script as well
	DefaultNativePrecompiles = []string{WEVMOSContractMainnet}
	// DefaultDynamicPrecompiles defines the default active dynamic precompiles
	DefaultDynamicPrecompiles []string
)

// NewParams creates a new Params object
func NewParams(
	enableErc20 bool,
	nativePrecompiles []string,
	dynamicPrecompiles []string,
) Params {
	slices.Sort(nativePrecompiles)
	slices.Sort(dynamicPrecompiles)
	return Params{
		EnableErc20:        enableErc20,
		NativePrecompiles:  nativePrecompiles,
		DynamicPrecompiles: dynamicPrecompiles,
	}
}

func DefaultParams() Params {
	return Params{
		EnableErc20:        true,
		NativePrecompiles:  DefaultNativePrecompiles,
		DynamicPrecompiles: DefaultDynamicPrecompiles,
	}
}

func ValidateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func (p Params) Validate() error {
	if err := ValidateBool(p.EnableErc20); err != nil {
		return err
	}

	if err := ValidatePrecompiles(p.NativePrecompiles); err != nil {
		return err
	}

	if err := ValidatePrecompiles(p.DynamicPrecompiles); err != nil {
		return err
	}

	combined := p.DynamicPrecompiles
	combined = append(combined, p.NativePrecompiles...)
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

	// NOTE: Check that the precompiles are sorted. This is required
	// to ensure determinism
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
