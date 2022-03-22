package types

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
)

// DefaultGenesis returns the default claims module genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:        DefaultParams(),
		ClaimsRecords: []ClaimsRecordAddress{},
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

	for _, claimsRecord := range gs.ClaimsRecords {
		if seenClaims[claimsRecord.Address] {
			return fmt.Errorf("duplicated claims record entry %s", claimsRecord.Address)
		}
		if err := claimsRecord.Validate(); err != nil {
			return err
		}
		seenClaims[claimsRecord.Address] = true
	}

	return gs.Params.Validate()
}
