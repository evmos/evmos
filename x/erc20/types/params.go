package types

import (
	fmt "fmt"
)

// Parameter store key
var (
	ParamStoreKeyEnableErc20   = []byte("EnableErc20")
	ParamStoreKeyEnableEVMHook = []byte("EnableEVMHook")
)

// NewParams creates a new Params object
func NewParams(
	enableErc20 bool,
	enableEVMHook bool,
) Params {
	return Params{
		EnableErc20:   enableErc20,
		EnableEVMHook: enableEVMHook,
	}
}

func DefaultParams() Params {
	return Params{
		EnableErc20:   true,
		EnableEVMHook: true,
	}
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func (p Params) Validate() error {
	if err := validateBool(p.EnableEVMHook); err != nil {
		return err
	}

	return validateBool(p.EnableErc20)
}
