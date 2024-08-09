// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package auctions

const (
	// ErrDifferentOriginFromSender is raised when the origin address is not the same as the delegator address.
	ErrDifferentOriginFromSender = "origin address %s is not the same as sender address %s"
	// ErrInvalidInputLength is raised when the input length is invalid.
	ErrInvalidInputLength = "invalid input length expected 0 got %d"
)
