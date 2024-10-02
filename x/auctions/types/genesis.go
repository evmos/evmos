// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesisState sets default auctions genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: Params{
			EnableAuction: true,
		},
		Bid: Bid{
			Sender:   "",
			BidValue: sdk.NewCoin(BidDenom, math.ZeroInt()),
		},
		Round: 0,
	}
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, bid Bid, round uint64) *GenesisState {
	return &GenesisState{
		Params: params,
		Bid:    bid,
		Round:  round,
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	if gs.Bid.BidValue.Denom != BidDenom {
		return errors.Wrapf(ErrInvalidDenom, "bid denom should be %s", BidDenom)
	}

	if gs.Bid.BidValue.IsNegative() {
		return errors.Wrapf(ErrNegativeBid, "bid amount should be positive")
	}

	return gs.Params.Validate()
}
