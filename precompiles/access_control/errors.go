// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package accesscontrol

const (
	// ErrInvalidMinterAddress is the error message for invalid minter address.
	ErrInvalidMinterAddress = "invalid minter address"
	// ErrInvalidRoleArgument is the error message for invalid role argument type.
	ErrInvalidRoleArgument = "invalid role argument type"
	// ErrInvalidAccountArgument is the error message for invalid account argument.
	ErrInvalidAccountArgument = "invalid account argument"
	// ErrBurnAmountNotGreaterThanZero is the error message for burn amount not greater than 0.
	ErrBurnAmountNotGreaterThanZero = "burn amount not greater than 0"
	// ErrMintAmountNotGreaterThanZero is the error message for mint amount not greater than 0.
	ErrMintAmountNotGreaterThanZero = "mint amount not greater than 0"
	// ErrMintToZeroAddress is the error message for mint to the zero address.
	ErrMintToZeroAddress = "attempting to mint to the zero address"
	// ErrSenderNoRole is the error message when sender does not have the role.
	ErrSenderNoRole = "access_control: sender does not have the role"
	// ErrRenounceRoleDifferentThanCaller is the error message when renouncing role different from the caller.
	ErrRenounceRoleDifferentThanCaller = "access_control: can only renounce roles for self"
)
