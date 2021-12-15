package types

import "fmt"

func NewGenesisState(params Params, incentives []Incentive) GenesisState {
	return GenesisState{
		Params:     params,
		Incentives: incentives,
	}
}

func (gs GenesisState) Validate() error {
	seenContract := make(map[string]bool)

	for _, b := range gs.Incentives {
		if seenContract[b.Contract] {
			return fmt.Errorf("contract duplicated on genesis '%s'", b.Contract)
		}

		if err := b.Validate(); err != nil {
			return err
		}

		seenContract[b.Contract] = true
	}

	return gs.Params.Validate()
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}
