package types

// NewGenesisState creates a new GenesisState object
func NewGenesisState(
	minter Minter,
	params Params,
	halvenStartedEpoch int64,
) GenesisState {
	return GenesisState{
		Minter:             minter,
		Params:             params,
		HalvenStartedEpoch: halvenStartedEpoch,
	}
}

// TODO do we need to set Minter and HalvenStarted Epoch here?
// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Minter:             DefaultInitialMinter(),
		Params:             DefaultParams(),
		HalvenStartedEpoch: 0,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	return gs.Minter.Validate()
}
