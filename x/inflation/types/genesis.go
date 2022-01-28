package types

import (
	fmt "fmt"

	epochtypes "github.com/tharsis/evmos/x/epochs/types"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(
	params Params,
	period uint64,
	epochIdentifier string,
	epochsPerPeriod int64,
) GenesisState {
	return GenesisState{
		Params:          params,
		Period:          period,
		EpochIdentifier: epochIdentifier,
		EpochsPerPeriod: epochsPerPeriod,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:          DefaultParams(),
		Period:          uint64(0),
		EpochIdentifier: "day",
		EpochsPerPeriod: 365,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := epochtypes.ValidateEpochIdentifierInterface(gs.EpochIdentifier); err != nil {
		return err
	}
	if err := validateEpochsPerPeriod(gs.EpochsPerPeriod); err != nil {
		return err
	}

	return gs.Params.Validate()
}

func validateEpochsPerPeriod(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("epochs per period must be positive: %d", v)
	}

	return nil
}
