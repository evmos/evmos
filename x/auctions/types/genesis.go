// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	epochstypes "github.com/evmos/evmos/v18/x/epochs/types"
	"time"
)

var (
	WeeklyDuration  = time.Hour * 24 * 7
	AuctionDuration = time.Hour * 24 * 4
)

// DefaultGenesisState sets default auctions genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: Params{
			EnableAuction: true,
		},
		AuctionEpoch: EpochInfo{
			Identifier:              epochstypes.WeekEpochID,
			StartTime:               time.Now(),
			Duration:                time.Hour * 24 * 7,
			CurrentEpoch:            0,
			CurrentEpochStartTime:   time.Now(),
			EpochCountingStarted:    true,
			CurrentEpochStartHeight: 0,
		},
	}
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, auctionEpoch EpochInfo) *GenesisState {
	return &GenesisState{
		Params:       params,
		AuctionEpoch: auctionEpoch,
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	return nil
}
