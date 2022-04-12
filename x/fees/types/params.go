package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store key
var (
	DefaultEnableFees      = false
	DefaultDeveloperShares = sdk.NewDecWithPrec(50, 2) // 50%
	DefaultValidatorShares = sdk.NewDecWithPrec(50, 2) // 50%
	// cost for `crypto.CreateAddress`
	// keccak256(word) costs 36 gas
	DefaultAddrDerivationCostCreate       = uint64(50)
	ParamStoreKeyEnableFees               = []byte("EnableFees")
	ParamStoreKeyDeveloperShares          = []byte("DeveloperShares")
	ParamStoreKeyValidatorShares          = []byte("ValidatorShares")
	ParamStoreKeyAddrDerivationCostCreate = []byte("AddrDerivationCostCreate")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params object
func NewParams(
	enableFees bool,
	developerShares,
	validatorShares sdk.Dec,
	addrDerivationCostCreate uint64,
) Params {
	return Params{
		EnableFees:               enableFees,
		DeveloperShares:          developerShares,
		ValidatorShares:          validatorShares,
		AddrDerivationCostCreate: addrDerivationCostCreate,
	}
}

func DefaultParams() Params {
	return Params{
		EnableFees:               DefaultEnableFees,
		DeveloperShares:          DefaultDeveloperShares,
		ValidatorShares:          DefaultValidatorShares,
		AddrDerivationCostCreate: DefaultAddrDerivationCostCreate,
	}
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableFees, &p.EnableFees, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyDeveloperShares, &p.DeveloperShares, validateShares),
		paramtypes.NewParamSetPair(ParamStoreKeyValidatorShares, &p.ValidatorShares, validateShares),
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

func (p Params) Validate() error {
	if err := validateBool(p.EnableFees); err != nil {
		return err
	}
	if err := validateShares(p.DeveloperShares); err != nil {
		return err
	}
	if err := validateShares(p.ValidatorShares); err != nil {
		return err
	}
	if p.DeveloperShares.Add(p.ValidatorShares).GT(sdk.OneDec()) {
		return fmt.Errorf("total shares cannot be greater than 1: %#s + %#s", p.DeveloperShares, p.ValidatorShares)
	}
	if err := validateUint64(p.AddrDerivationCostCreate); err != nil {
		return err
	}

	return nil
}
