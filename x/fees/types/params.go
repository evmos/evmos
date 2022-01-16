package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store key
var (
	ParamStoreKeyContractDistribution  = []byte("ContractDistribution")
	ParamStoreKeyEnableWithdrawAddress = []byte("WithdrawAddress")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params object
func NewParams(
	enableWithdrawAddress bool,
	distr sdk.Dec,
) Params {
	return Params{
		WithdrawAddrEnabled:  enableWithdrawAddress,
		ContractDistribution: distr,
	}
}

// DefaultParams returns a Params instance the default module parameter values
func DefaultParams() Params {
	return Params{
		WithdrawAddrEnabled:  true,
		ContractDistribution: sdk.NewDecWithPrec(5, 1), // 50%
	}
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyContractDistribution, &p.ContractDistribution, validateContractDistribution),
		paramtypes.NewParamSetPair(ParamStoreKeyEnableWithdrawAddress, &p.WithdrawAddrEnabled, validateBool),
	}
}

// Validate performs a stateless validation of the distribution fields
func (p Params) Validate() error {
	return validateContractDistribution(p.ContractDistribution)
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateContractDistribution(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("distribution cannot be negative: %s", v)
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("distribution cannot be > 1: %s", v)
	}

	return nil
}
