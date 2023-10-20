// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package osmosis

const (
	OutpostName = "osmosis-outpost"
)

var (
	ErrSlippagePercentage = "slippage percentage must be a string containing an uint64 type"
	// ErrTokenPairNotFound is raised when a token pair for a certain address
	// is not found and it is required by the executing function.
	ErrTokenPairNotFound = "token pair for address %s not found"
	// ErrInputTokenNotSupported is raised when a the osmosis outpost receive a non supported
	// input token for the swap.
	ErrInputTokenNotSupported = "input not supported, supported tokens: %v"
)
