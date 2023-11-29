// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package osmosis

import "fmt"

var (
	// ErrEmptyOnFailedDelivery is raised when the onFailedDeliver field of the
	// IBC memo is an empty string.
	ErrEmptyOnFailedDelivery = "onFailedDelivery cannot be empty"
	// ErrInputEqualOutput is raised when input and output tokens are the same.
	ErrInputEqualOutput = "input and output token cannot be the same: %s"
	// ErrSlippagePercentage is raised when the requested slippage percentage is
	// higher than a pre-defined amount.
	ErrSlippagePercentage = fmt.Sprintf("slippage percentage must be: 0 < slippagePercentage <= %d", MaxSlippagePercentage)
	// ErrWindowSeconds is raised when the requested window seconds is
	// higher than a pre-defined amount.
	ErrWindowSeconds = fmt.Sprintf("window seconds must be: 0 < windowSeconds <= %d", MaxWindowSeconds)
	// ErrTokenPairNotFound is raised when a token pair for a certain address
	// is not found, and it is required by the executing function.
	ErrTokenPairNotFound = "token pair for address %s not found"
	// ErrInputTokenNotSupported is raised when the osmosis outpost receives a non-supported
	// input token for the swap.
	ErrInputTokenNotSupported = "input not supported, supported tokens: %v" //#nosec G101 -- no hardcoded credentials here
	// ErrReceiverAddress is raised when an error occurs during the validation of the receiver of the
	// swap.
	ErrReceiverAddress = "error during receiver address validation: %s"
)
