package types

import (
	"encoding/json"
	"fmt"
	time "time"

	"github.com/cosmos/cosmos-sdk/codec"
)

// DefaultGenesis returns the default claims module genesis state
func DefaultGenesis(airdropStartTime time.Time) *GenesisState {
	return &GenesisState{
		Params:       DefaultParams(airdropStartTime),
		ClaimRecords: []ClaimRecordAddress{},
	}
}

// GetGenesisStateFromAppState returns x/claims GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	seenClaims := make(map[string]bool)

	for _, claimRecord := range gs.ClaimRecords {
		if seenClaims[claimRecord.Address] {
			return fmt.Errorf("duplicated claim record entry %s", claimRecord.Address)
		}
		if err := claimRecord.Validate(); err != nil {
			return err
		}
		seenClaims[claimRecord.Address] = true
	}

	return gs.Params.Validate()
}
