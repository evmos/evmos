package types

import (
	"time"

	"github.com/tharsis/evmos/v4/x/epochs/types"
)

// DefaultGenesisState returns the default epochs genesis state for v1
func DefaultGenesisState() *GenesisState {
	epochs := []EpochInfo{
		{
			Identifier:              types.WeekEpochID,
			StartTime:               time.Time{},
			Duration:                time.Hour * 24 * 7,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
		{
			Identifier:              types.DayEpochID,
			StartTime:               time.Time{},
			Duration:                time.Hour * 24,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
	}
	return &GenesisState{Epochs: epochs}
}
