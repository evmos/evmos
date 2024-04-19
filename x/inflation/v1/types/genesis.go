// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	fmt "fmt"

	epochstypes "github.com/evmos/evmos/v17/x/epochs/types"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(
	params Params,
	period uint64,
	epochIdentifier string,
	epochsPerPeriod int64,
	skippedEpochs uint64,
) GenesisState {
	return GenesisState{
		Params:          params,
		Period:          period,
		EpochIdentifier: epochIdentifier,
		EpochsPerPeriod: epochsPerPeriod,
		SkippedEpochs:   skippedEpochs,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:          DefaultParams(),
		Period:          uint64(0),
		EpochIdentifier: epochstypes.DayEpochID,
		EpochsPerPeriod: 365,
		SkippedEpochs:   0,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := epochstypes.ValidateEpochIdentifierInterface(gs.EpochIdentifier); err != nil {
		return err
	}

	if err := validateEpochsPerPeriod(gs.EpochsPerPeriod); err != nil {
		return err
	}

	if err := validateSkippedEpochs(gs.SkippedEpochs); err != nil {
		return err
	}

	return gs.Params.Validate()
}

func validateEpochsPerPeriod(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid genesis state type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("epochs per period must be positive: %d", v)
	}

	return nil
}

func validateSkippedEpochs(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid genesis state type: %T", i)
	}
	return nil
}
