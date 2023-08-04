// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"fmt"
)

var ParamsKey = []byte("Params")

var DefaultEnableGovClawback = true

// NewParams creates a new Params object
func NewParams(enableGovClawback bool) Params {
	return Params{
		EnableGovClawback: enableGovClawback,
	}
}

// DefaultParams default vesting module parameters
func DefaultParams() Params {
	return Params{
		EnableGovClawback: DefaultEnableGovClawback,
	}
}

// Validate validates the params and returns error if any
func (p Params) Validate() error {
	return validateBool(p.EnableGovClawback)
}

// validateBool validates bool params
func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
