package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Actions []Action

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:       DefaultParams(),
		ClaimRecords: []ClaimRecord{},
	}
}

// GetGenesisStateFromAppState returns x/claims GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// TODO: validate records
	for _, claimRecord := range gs.ClaimRecords {
		if _, err := sdk.AccAddressFromBech32(claimRecord.Address); err != nil {
			return err
		}
		if claimRecord.InitialClaimableAmount.Empty() {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "empty coins")
		}
		if err := claimRecord.InitialClaimableAmount.Validate(); err != nil {
			return err
		}
		// TODO: check repeated actions
	}

	return gs.Params.Validate()
}
