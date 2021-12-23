package types

import (
	"errors"
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store key
var (
	ParamStoreKeyEnableIncentives = []byte("EnableIncentives")
	ParamStoreKeyEpochDuration    = []byte("EpochDuration")
	ParamStoreKeyAllocationLimit  = []byte("AllocationLimit")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params object
func NewParams(
	enableIncentives bool,
	epocheDuration time.Duration,
	allocationLimit sdk.Dec,

) Params {
	return Params{
		EnableIncentives: enableIncentives,
		EpochDuration:    epocheDuration,
		AllocationLimit:  allocationLimit,
	}
}

func DefaultParams() Params {
	return Params{
		EnableIncentives: true,
		EpochDuration:    govtypes.DefaultPeriod,
		AllocationLimit:  sdk.NewDecWithPrec(5, 2),
	}
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableIncentives, &p.EnableIncentives, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyEpochDuration, &p.EpochDuration, validatePeriod),
		paramtypes.NewParamSetPair(ParamStoreKeyAllocationLimit, &p.AllocationLimit, validatePercentage),
	}
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validatePeriod(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("voting period must be positive: %s", v)
	}

	return nil
}

func validatePercentage(i interface{}) error {
	dec, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if dec.IsNil() {
		return errors.New("allocation limit cannot be nil")
	}
	if dec.IsNegative() {
		return fmt.Errorf("allocation limit must be positive: %s", dec)
	}
	if dec.GT(sdk.OneDec()) {
		return fmt.Errorf("allocation limit must <= 100: %s", dec)
	}

	return nil
}

func (p Params) Validate() error {
	if err := validateBool(p.EnableIncentives); err != nil {
		return err
	}

	if err := validatePeriod(p.EpochDuration); err != nil {
		return err
	}

	if err := validatePercentage(p.AllocationLimit); err != nil {
		return err
	}

	return nil
}
