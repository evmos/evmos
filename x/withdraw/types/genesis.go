package types

import "fmt"

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, prefixes ...Prefix) GenesisState {
	return GenesisState{
		Params:   params,
		Prefixes: prefixes,
	}
}

// DefaultGenesisState sets default evm genesis state with empty accounts and
// default params and chain config values.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:   DefaultParams(),
		Prefixes: []Prefix{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	seenPrefixes := make(map[string]bool)
	for _, prefix := range gs.Prefixes {
		if seenPrefixes[prefix.SourceChannel] {
			return fmt.Errorf("duplicated source channel %s", prefix.SourceChannel)
		}

		if err := prefix.Validate(); err != nil {
			return err
		}

		seenPrefixes[prefix.SourceChannel] = true
	}

	return gs.Params.Validate()
}
