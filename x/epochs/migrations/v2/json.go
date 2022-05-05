package v2

import (
	"time"

	"github.com/tharsis/evmos/v4/x/epochs/types"
)

var NewEpochs = []types.EpochInfo{
	{
		Identifier:              YearEpochID,
		StartTime:               time.Time{},
		Duration:                time.Hour * 24 * 365,
		CurrentEpoch:            0,
		CurrentEpochStartHeight: 0,
		CurrentEpochStartTime:   time.Time{},
		EpochCountingStarted:    false,
	},
	{
		Identifier:              HourEpochID,
		StartTime:               time.Time{},
		Duration:                time.Hour,
		CurrentEpoch:            0,
		CurrentEpochStartHeight: 0,
		CurrentEpochStartTime:   time.Time{},
		EpochCountingStarted:    false,
	},
}

// MigrateJSON accepts exported 1 x/epochs genesis state and migrates it
// to 2 x/epochs genesis state. Hourly and yearly epochs are added.
func MigrateJSON(state types.GenesisState) types.GenesisState {
	state.Epochs = append(state.Epochs, NewEpochs[:]...)
	return state
}
