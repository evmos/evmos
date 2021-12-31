package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/tharsis/evmos/x/epochs/types"
)

// RandomizedGenState generates a random GenesisState for mint
func RandomizedGenState(simState *module.SimulationState) {
	epochs := []types.EpochInfo{
		{
			Identifier:              "day",
			StartTime:               time.Time{},
			Duration:                time.Hour * 24,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
		{
			Identifier:              "hour",
			StartTime:               time.Time{},
			Duration:                time.Hour,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
	}
	epochGenesis := types.NewGenesisState(epochs)

	bz, err := json.MarshalIndent(&epochGenesis, "", " ")
	if err != nil {
		panic(err)
	}

	// TODO: Do some randomization later
	fmt.Printf("Selected deterministically generated epoch parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(epochGenesis)
}
