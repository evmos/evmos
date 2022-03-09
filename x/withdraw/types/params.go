package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store key
var (
	ParamStoreKeyEnableWithdraw = []byte("EnableWithdraw")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	enableWithdraw bool,
	enabledChannels ...string,
) Params {
	return Params{
		EnableWithdraw: enableWithdraw,
	}
}

// DefaultParams defines the default params for the withdraw module
func DefaultParams() Params {
	return Params{
		EnableWithdraw: true,
	}
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableWithdraw, &p.EnableWithdraw, validateBool),
	}
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

// Validate checks that the fields have valid values
func (p Params) Validate() error {
	return nil
}
