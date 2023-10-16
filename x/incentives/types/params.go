// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/evmos/evmos/v15/x/epochs/types"
)

// ParamsKey params store key
var ParamsKey = []byte("Params")

var (
	DefaultEnableIncentives          = true
	DefaultAllocationLimit           = sdk.NewDecWithPrec(5, 2)
	DefaultIncentivesEpochIdentifier = epochstypes.WeekEpochID
	DefaultRewardScalar              = sdk.NewDecWithPrec(12, 1)
)

// NewParams creates a new Params object
func NewParams(
	enableIncentives bool,
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
		EnableIncentives:          DefaultEnableIncentives,
		AllocationLimit:           DefaultAllocationLimit,
		IncentivesEpochIdentifier: DefaultIncentivesEpochIdentifier,
		RewardScaler:              DefaultRewardScalar,
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
