// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package vesting

const (
	// ErrDifferentFromOrigin is raised when the tx origin address is not the same as the vesting transaction initiator.
	ErrDifferentFromOrigin = "tx origin address %s does not match the from address %s"
	// ErrDifferentFunderOrigin is raised when the tx origin address is not the same as the vesting transaction funder.
	ErrDifferentFunderOrigin = "tx origin address %s does not match the funder address %s"
)
