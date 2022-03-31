package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store key
var (
	DefaultDeveloperPercentage       = sdk.NewDecWithPrec(50, 2) // 50%
	DefaultValidatorPercentage       = sdk.NewDecWithPrec(50, 2) // 50%
	ParamStoreKeyEnableFees          = []byte("EnableFees")
	ParamStoreKeyDeveloperPercentage = []byte("DeveloperPercentage")
	ParamStoreKeyValidatorPercentage = []byte("ValidatorPercentage")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params object
func NewParams(
	enableFees bool,
	developerPercentage sdk.Dec,
	validatorPercentage sdk.Dec,

) Params {
	return Params{
		EnableFees:          enableFees,
		DeveloperPercentage: developerPercentage,
		ValidatorPercentage: validatorPercentage,
	}
}

func DefaultParams() Params {
	return Params{
		EnableFees:          true,
		DeveloperPercentage: DefaultDeveloperPercentage,
		ValidatorPercentage: DefaultValidatorPercentage,
	}
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableFees, &p.EnableFees, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyDeveloperPercentage, &p.DeveloperPercentage, validatePercentage),
		paramtypes.NewParamSetPair(ParamStoreKeyValidatorPercentage, &p.ValidatorPercentage, validatePercentage),
	}
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validatePercentage(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("value cannot be negative")
	}

	return nil
}

func (p Params) Validate() error {
	if err := validateBool(p.EnableFees); err != nil {
		return err
	}
	if err := validatePercentage(p.DeveloperPercentage); err != nil {
		return err
	}
	if err := validatePercentage(p.ValidatorPercentage); err != nil {
		return err
	}

	return nil
}
