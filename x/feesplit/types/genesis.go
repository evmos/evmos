package types

import "fmt"

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, feeSplits []FeeSplit) GenesisState {
	return GenesisState{
		Params:    params,
		FeeSplits: feeSplits,
	}
}

// DefaultGenesisState sets default evm genesis state with empty accounts and
// default params and chain config values.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	seenContract := make(map[string]bool)
	for _, fs := range gs.FeeSplits {
		// only one fee per contract
		if seenContract[fs.ContractAddress] {
			return fmt.Errorf("contract duplicated on genesis '%s'", fs.ContractAddress)
		}

		if err := fs.Validate(); err != nil {
			return err
		}

		seenContract[fs.ContractAddress] = true
	}

	return gs.Params.Validate()
}
