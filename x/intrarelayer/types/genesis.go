package types

import "fmt"

func NewGenesisState(params Params, pairs []TokenPair) GenesisState {
	return GenesisState{
		Params:     params,
		TokenPairs: pairs,
	}
}

func (gs GenesisState) Validate() error {
	seenTokenPairs := make(map[string]bool)

	for _, b := range gs.TokenPairs {
		id := b.Erc20Address + "|" + b.Denom
		if seenTokenPairs[id] {
			return fmt.Errorf("duplicate token map '%s'", id)
		}

		if err := b.Validate(); err != nil {
			return err
		}

		seenTokenPairs[id] = true
	}

	return gs.Params.Validate()
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}
