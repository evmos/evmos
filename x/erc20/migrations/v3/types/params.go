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

package v3types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/evmos/evmos/v12/x/erc20/types"
)

var _ types.LegacyParams = &V3Params{}

var (
	DefaultErc20   = true
	DefaultEVMHook = true
)

// Parameter store key
var (
	ParamStoreKeyEnableErc20   = []byte("EnableErc20")
	ParamStoreKeyEnableEVMHook = []byte("EnableEVMHook")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&V3Params{})
}

// ParamSetPairs returns the parameter set pairs.
func (p *V3Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableErc20, &p.EnableErc20, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyEnableEVMHook, &p.EnableEVMHook, validateBool),
	}
}

// NewParams creates a new Params object
func NewParams(
	enableErc20 bool,
	enableEVMHook bool,
) V3Params {
	return V3Params{
		EnableErc20:   enableErc20,
		EnableEVMHook: enableEVMHook,
	}
}

func DefaultParams() V3Params {
	return V3Params{
		EnableErc20:   DefaultErc20,
		EnableEVMHook: DefaultEVMHook,
	}
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func (p V3Params) Validate() error {
	if err := validateBool(p.EnableEVMHook); err != nil {
		return err
	}

	return validateBool(p.EnableErc20)
}
