package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Actions []Action

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		ModuleAccountBalance: sdk.NewCoin(DefaultClaimDenom, sdk.ZeroInt()),
		Params: Params{
			AirdropStartTime:   time.Time{},
			DurationUntilDecay: DefaultDurationUntilDecay, // 2 month
			DurationOfDecay:    DefaultDurationOfDecay,    // 4 months
			ClaimDenom:         DefaultClaimDenom,         // uosmo
		},
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
	totalClaimable := sdk.Coins{}

	for _, claimRecord := range gs.ClaimRecords {
		totalClaimable = totalClaimable.Add(claimRecord.InitialClaimableAmount...)
	}

	// if !totalClaimable.IsEqual(sdk.NewCoins(gs.ModuleAccountBalance)) {
	// 	return ErrIncorrectModuleAccountBalance
	// }

	if gs.Params.ClaimDenom != gs.ModuleAccountBalance.Denom {
		return fmt.Errorf("denom for module and claim does not match")
	}

	return nil
}
