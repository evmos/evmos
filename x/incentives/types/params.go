package types

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	epochstypes "github.com/tharsis/evmos/v5/x/epochs/types"
)

// Parameter store key
var (
	ParamStoreKeyEnableIncentives = []byte("EnableIncentives")
	ParamStoreKeyAllocationLimit  = []byte("AllocationLimit")
	ParamStoreKeyEpochIdentifier  = []byte("EpochIdentifier")
	ParamStoreKeyRewardScaler     = []byte("RewardScaler")
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
	epochIdentifier string,
	rewardScaler sdk.Dec,
) Params {
	return Params{
		EnableIncentives:          enableIncentives,
		AllocationLimit:           allocationLimit,
		IncentivesEpochIdentifier: epochIdentifier,
		RewardScaler:              rewardScaler,
	}
}

func DefaultParams() Params {
	return Params{
		EnableIncentives:          true,
		AllocationLimit:           sdk.NewDecWithPrec(5, 2),
		IncentivesEpochIdentifier: epochstypes.WeekEpochID,
		RewardScaler:              sdk.NewDecWithPrec(12, 1),
	}
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnableIncentives, &p.EnableIncentives, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyAllocationLimit, &p.AllocationLimit, validatePercentage),
		paramtypes.NewParamSetPair(ParamStoreKeyEpochIdentifier, &p.IncentivesEpochIdentifier, epochstypes.ValidateEpochIdentifierInterface),
		paramtypes.NewParamSetPair(ParamStoreKeyRewardScaler, &p.RewardScaler, validateUncappedPercentage),
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

func validateUncappedPercentage(i interface{}) error {
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

	return nil
}

func (p Params) Validate() error {
	if err := validateBool(p.EnableIncentives); err != nil {
		return err
	}

	if err := validatePercentage(p.AllocationLimit); err != nil {
		return err
	}

	if err := validateUncappedPercentage(p.RewardScaler); err != nil {
		return err
	}

	return epochstypes.ValidateEpochIdentifierString(p.IncentivesEpochIdentifier)
}
