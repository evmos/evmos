package types

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, period int64) GenesisState {
	return GenesisState{
		Params: params,
		Period: period,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
		Period: int64(0),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}
