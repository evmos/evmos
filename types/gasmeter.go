// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package types

import (
	fmt "fmt"
	math "math"

	storetypes "cosmossdk.io/store/types"
)

// ErrorNegativeGasConsumed defines an error thrown when the amount of gas refunded results in a
// negative gas consumed amount.
// Copied from cosmos-sdk
type ErrorNegativeGasConsumed struct {
	Descriptor string
}

// ErrorGasOverflow defines an error thrown when an action results gas consumption
// unsigned integer overflow.
type ErrorGasOverflow struct {
	Descriptor string
}

type infiniteGasMeterWithLimit struct {
	consumed storetypes.Gas
	limit    storetypes.Gas
}

// NewInfiniteGasMeterWithLimit returns a reference to a new infiniteGasMeter.
func NewInfiniteGasMeterWithLimit(limit storetypes.Gas) storetypes.GasMeter {
	return &infiniteGasMeterWithLimit{
		consumed: 0,
		limit:    limit,
	}
}

// GasConsumed returns the gas consumed from the GasMeter.
func (g *infiniteGasMeterWithLimit) GasConsumed() storetypes.Gas {
	return g.consumed
}

// GasConsumedToLimit returns the gas limit if gas consumed is past the limit,
// otherwise it returns the consumed gas.
// NOTE: This behavior is only called when recovering from panic when
// BlockGasMeter consumes gas past the limit.
func (g *infiniteGasMeterWithLimit) GasConsumedToLimit() storetypes.Gas {
	return g.consumed
}

// Limit returns the gas limit of the GasMeter.
func (g *infiniteGasMeterWithLimit) Limit() storetypes.Gas {
	return g.limit
}

// addUint64Overflow performs the addition operation on two uint64 integers and
// returns a boolean on whether or not the result overflows.
func addUint64Overflow(a, b uint64) (uint64, bool) {
	if math.MaxUint64-a < b {
		return 0, true
	}

	return a + b, false
}

// ConsumeGas adds the given amount of gas to the gas consumed and panics if it overflows the limit or out of gas.
func (g *infiniteGasMeterWithLimit) ConsumeGas(amount storetypes.Gas, descriptor string) {
	var overflow bool
	// TODO: Should we set the consumed field after overflow checking?
	g.consumed, overflow = addUint64Overflow(g.consumed, amount)
	if overflow {
		panic(ErrorGasOverflow{descriptor})
	}
}

// RefundGas will deduct the given amount from the gas consumed. If the amount is greater than the
// gas consumed, the function will panic.
//
// Use case: This functionality enables refunding gas to the transaction or block gas pools so that
// EVM-compatible chains can fully support the go-ethereum StateDb interface.
// See https://github.com/cosmos/cosmos-sdk/pull/9403 for reference.
func (g *infiniteGasMeterWithLimit) RefundGas(amount storetypes.Gas, descriptor string) {
	if g.consumed < amount {
		panic(ErrorNegativeGasConsumed{Descriptor: descriptor})
	}

	g.consumed -= amount
}

// IsPastLimit returns true if gas consumed is past limit, otherwise it returns false.
func (g *infiniteGasMeterWithLimit) IsPastLimit() bool {
	return false
}

// IsOutOfGas returns true if gas consumed is greater than or equal to gas limit, otherwise it returns false.
func (g *infiniteGasMeterWithLimit) IsOutOfGas() bool {
	return false
}

// String returns the BasicGasMeter's gas limit and gas consumed.
func (g *infiniteGasMeterWithLimit) String() string {
	return fmt.Sprintf("InfiniteGasMeter:\n  consumed: %d", g.consumed)
}

// GasRemaining returns MaxUint64 since limit is not confined in infiniteGasMeter.
func (g *infiniteGasMeterWithLimit) GasRemaining() storetypes.Gas {
	return math.MaxUint64
}
