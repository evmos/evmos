// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

// NewGenesisState creates a new genesis state instance
func NewGenesisState(params Params, data []FlaggedInfo) *GenesisState {
	return &GenesisState{
		Params:      params,
		FlaggedData: data,
	}
}

// DefaultGenesisState returns the default epochs genesis state
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), []FlaggedInfo{})
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// TODO: add validation
	return nil
}
