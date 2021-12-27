package types

import (
	"errors"
	"time"
)

func NewGenesisState(epochs []EpochInfo) *GenesisState {
	return &GenesisState{Epochs: epochs}
}

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	epochs := []EpochInfo{
		{
			Identifier:              "week",
			StartTime:               time.Time{},
			Duration:                time.Hour * 24 * 7,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
		{
			Identifier:              "day",
			StartTime:               time.Time{},
			Duration:                time.Hour * 24,
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
	epochIdentifiers := map[string]bool{}
	for _, epoch := range gs.Epochs {
		if epoch.Identifier == "" {
			return errors.New("epoch identifier should NOT be empty")
		}
		if epochIdentifiers[epoch.Identifier] {
			return errors.New("epoch identifier should be unique")
		}
		if epoch.Duration == 0 {
			return errors.New("epoch duration should NOT be 0")
		}
		epochIdentifiers[epoch.Identifier] = true
	}
	return nil
}
