package types

func NewGenesisState(params Params) GenesisState {
	return GenesisState{
		Params: params,
	}
}

func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}
