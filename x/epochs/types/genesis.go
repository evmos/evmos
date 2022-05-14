package types

import (
	"fmt"
	"time"
)

// NewGenesisState creates a new genesis state instance
func NewGenesisState(epochs []EpochInfo) *GenesisState {
	return &GenesisState{Epochs: epochs}
}

// DefaultGenesisState returns the default epochs genesis state
// Changing genesis state requires updating the IdentifierToDuration
// and DurationToIdentifier maps
func DefaultGenesisState() *GenesisState {
	epochs := []EpochInfo{
		{
			StartTime:               time.Time{},
			Duration:                WeekEpochDuration,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
		{
			StartTime:               time.Time{},
			Duration:                DayEpochDuration,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
	}
	return NewGenesisState(epochs)
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	epochIdentifiers := make(map[string]bool)

	for _, epoch := range gs.Epochs {
		if epochIdentifiers[epoch.Duration.String()] {
			return fmt.Errorf("duplicated epoch entry %s", DurationToIdentifier[epoch.Duration])
		}
		if err := epoch.Validate(); err != nil {
			return err
		}
		epochIdentifiers[epoch.Duration.String()] = true
	}

	return nil
}
