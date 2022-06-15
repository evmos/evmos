package types

import "fmt"

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, fees []Fee) GenesisState {
	return GenesisState{
		Params: params,
		Fees:   fees,
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
	seenContractIn := make(map[string]bool)
	for _, in := range gs.Fees {
		// only one fee per contract
		if seenContractIn[in.ContractAddress] {
			return fmt.Errorf("contract duplicated on genesis '%s'", in.ContractAddress)
		}

		if err := in.Validate(); err != nil {
			return err
		}

		seenContractIn[in.ContractAddress] = true
	}

	return gs.Params.Validate()
}
