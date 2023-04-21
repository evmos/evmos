// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package types

// DefaultGenesisState sets default fee market genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:   DefaultParams(),
		BlockGas: 0,
	}
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, blockGas uint64) *GenesisState {
	return &GenesisState{
		Params:   params,
		BlockGas: blockGas,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}
