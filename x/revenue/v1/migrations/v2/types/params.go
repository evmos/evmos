// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package types

import (
	"fmt"

	"github.com/evmos/evmos/v12/x/revenue/v1/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ types.LegacyParams = &V2Params{}

// Parameter store key
var (
	DefaultEnableRevenue   = true
	DefaultDeveloperShares = sdk.NewDecWithPrec(50, 2) // 50%
	// DefaultAddrDerivationCostCreate Cost for executing `crypto.CreateAddress` must be at least 36 gas for the
	// contained keccak256(word) operation
	DefaultAddrDerivationCostCreate = uint64(50)
)

var (
	ParamsKey                             = []byte("Params")
	ParamStoreKeyEnableRevenue            = []byte("EnableRevenue")
	ParamStoreKeyDeveloperShares          = []byte("DeveloperShares")
	ParamStoreKeyAddrDerivationCostCreate = []byte("AddrDerivationCostCreate")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&V2Params{})
}

// NewParams creates a new Params object
func NewParams(
	enableRevenue bool,
	developerShares sdk.Dec,
	addrDerivationCostCreate uint64,
) V2Params {
	return V2Params{
		EnableRevenue:            enableRevenue,
		DeveloperShares:          developerShares,
		AddrDerivationCostCreate: addrDerivationCostCreate,
	}
}

func DefaultParams() V2Params {
	return V2Params{
		EnableRevenue:            DefaultEnableRevenue,
		DeveloperShares:          DefaultDeveloperShares,
		AddrDerivationCostCreate: DefaultAddrDerivationCostCreate,
	}
}

// ParamSetPairs returns the parameter set pairs.
func (p *V2Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableRevenue, &p.EnableRevenue, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyDeveloperShares, &p.DeveloperShares, validateShares),
		paramtypes.NewParamSetPair(ParamStoreKeyAddrDerivationCostCreate, &p.AddrDerivationCostCreate, validateUint64),
	}
}

func validateUint64(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateShares(i interface{}) error {
	v, ok := i.(sdk.Dec)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("invalid parameter: nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("value cannot be negative: %T", i)
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("value cannot be greater than 1: %T", i)
	}

	return nil
}

func (p V2Params) Validate() error {
	if err := validateBool(p.EnableRevenue); err != nil {
		return err
	}
	if err := validateShares(p.DeveloperShares); err != nil {
		return err
	}
	return validateUint64(p.AddrDerivationCostCreate)
}
