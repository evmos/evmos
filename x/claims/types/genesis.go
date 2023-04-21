// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

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
