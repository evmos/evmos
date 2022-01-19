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

// TODO replace ValidateGenesis with Validate method
// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	return ValidateMinter(data.Minter)
}
