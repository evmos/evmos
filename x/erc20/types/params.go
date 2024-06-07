// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"fmt"
)

// Parameter store key
var (
	ParamStoreKeyEnableErc20 = []byte("EnableErc20")
)

// NewParams creates a new Params object
func NewParams(
	enableErc20 bool,
) Params {
	return Params{
		EnableErc20: enableErc20,
	}
}

func DefaultParams() Params {
	return Params{
		EnableErc20: true,
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
    return nil
}

