// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package staking

const (
	// ErrDecreaseAmountTooBig is raised when the amount by which the allowance should be decreased is greater
	// than the authorization limit.
	ErrDecreaseAmountTooBig = "amount by which the allowance should be decreased is greater than the authorization limit: %s > %s"
	// ErrDifferentOriginFromDelegator is raised when the origin address is not the same as the delegator address.
	ErrDifferentOriginFromDelegator = "origin address %s is not the same as delegator address %s"
	// ErrNoDelegationFound is raised when no delegation is found for the given delegator and validator addresses.
	ErrNoDelegationFound = "delegation with delegator %s not found for validator %s"
)
