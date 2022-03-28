package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store key
var (
	DefaultFeesDenom                 = "adevi"
	DefaultDeveloperPercentage       = uint64(50)
	DefaultValidatorPercentage       = uint64(50)
	ParamStoreKeyEnableFees          = []byte("EnableFees")
	ParamStoreKeyFeesDenom           = []byte("FeesDenom")
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
	feesDenom string,
	developerPercentage uint64,
	validatorPercentage uint64,

) Params {
	return Params{
		EnableFees:          enableFees,
		FeesDenom:           feesDenom,
		DeveloperPercentage: developerPercentage,
		ValidatorPercentage: validatorPercentage,
	}
}

func DefaultParams() Params {
	return Params{
		EnableFees:          true,
		FeesDenom:           DefaultFeesDenom,
		DeveloperPercentage: DefaultDeveloperPercentage,
		ValidatorPercentage: DefaultValidatorPercentage,
	}
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableFees, &p.EnableFees, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyFeesDenom, &p.FeesDenom, validateDenom),
		paramtypes.NewParamSetPair(ParamStoreKeyDeveloperPercentage, &p.DeveloperPercentage, validateUint64),
		paramtypes.NewParamSetPair(ParamStoreKeyValidatorPercentage, &p.ValidatorPercentage, validateUint64),
	}
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateDenom(i interface{}) error {
	denom, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return sdk.ValidateDenom(denom)
}

func validateUint64(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func (p Params) Validate() error {
	if err := validateBool(p.EnableFees); err != nil {
		return err
	}
	if err := sdk.ValidateDenom(p.FeesDenom); err != nil {
		return err
	}
	if err := validateUint64(p.DeveloperPercentage); err != nil {
		return err
	}
	if err := validateUint64(p.ValidatorPercentage); err != nil {
		return err
	}

	return nil
}
