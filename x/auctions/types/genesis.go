// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

// DefaultGenesisState sets default auctions genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: Params{
			EnableAuction: true,
		},
	}
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params) *GenesisState {
	return &GenesisState{
		Params: params,
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	return nil
}
