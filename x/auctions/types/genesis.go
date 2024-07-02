// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import "time"

var (
	WeeklyDuration  = time.Hour * 24 * 7
	AuctionDuration = time.Hour * 24 * 4
)

// DefaultGenesisState sets default fee market genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: Params{
			EnableAuction: true,
		},
		AuctionStartEpoch: EpochInfo{
			Identifier:              "weekly_auction_start",
			StartTime:               time.Now(),
			Duration:                time.Hour * 24 * 7,
			CurrentEpoch:            0,
			CurrentEpochStartTime:   time.Now(),
			EpochCountingStarted:    true,
			CurrentEpochStartHeight: 0,
		},
		AuctionEndEpoch: EpochInfo{
			Identifier:              "weekly_auction_end",
			StartTime:               time.Now().Add(time.Hour * 24 * 3),
			Duration:                time.Hour * 24 * 7,
			CurrentEpoch:            0,
			CurrentEpochStartTime:   time.Now().Add(time.Hour * 24 * 3),
			EpochCountingStarted:    true,
			CurrentEpochStartHeight: 0,
		},
	}
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, auctionStartEpoch, auctionEndEpoch EpochInfo) *GenesisState {
	return &GenesisState{
		Params:            params,
		AuctionStartEpoch: auctionStartEpoch,
		AuctionEndEpoch:   auctionEndEpoch,
	}
}
