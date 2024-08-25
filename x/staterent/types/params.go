// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package types

import (
	"cosmossdk.io/math"
)

var (
	DefaultAmountOfEntriesNeededToBeFlagged = int64(1000)
	DefaultCurrentPrice                     = math.NewInt(0)
	DefaultNextPrice                        = math.NewInt(0)
	DefaultEntriesToDeletePerBlock          = int64(1000)
	DefaultCurrentTic                       = uint64(0)
	DefaultBlocksPerTic                     = uint64(1000)
	DefaultCurrentTicBlock                  = uint64(0)
)

// NewParams creates a new Params instance
func NewParams(
	entries int64,
	currentPrice math.Int,
	nextPrice math.Int,
	deletePerTick int64,
	currentTic uint64,
	blocksPerTic uint64,
	currentTicBlock uint64,

) Params {
	return Params{
		AmountOfEntriesNeededToBeFlagged: entries,
		CurrentPrice:                     currentPrice,
		NextPrice:                        nextPrice,
		EntriesToDeletePerBlock:          entries,
		CurrentTic:                       currentTic,
		BlocksPerTic:                     blocksPerTic,
		CurrentTicBlock:                  currentTicBlock,
	}
}

// DefaultParams returns default evm parameters
func DefaultParams() Params {
	return Params{
		AmountOfEntriesNeededToBeFlagged: DefaultAmountOfEntriesNeededToBeFlagged,
		CurrentPrice:                     DefaultCurrentPrice,
		NextPrice:                        DefaultNextPrice,
		EntriesToDeletePerBlock:          DefaultEntriesToDeletePerBlock,
		CurrentTic:                       DefaultCurrentTic,
		BlocksPerTic:                     DefaultBlocksPerTic,
		CurrentTicBlock:                  DefaultCurrentTicBlock,
	}
}

// Validate performs basic validation on evm parameters.
func (p Params) Validate() error {
	// TODO: add validation
	return nil
}
