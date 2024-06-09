// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package accesscontrol

const (
	// ErrorInvalidArgumentNumber is the error message for invalid number of arguments while parsing.
	ErrorInvalidArgumentNumber = "invalid number of arguments"
	// ErrorInvalidMinterAddress is the error message for invalid minter address.
	ErrorInvalidMinterAddress = "invalid minter address"
	// ErrorInvalidRoleArgument is the error message for invalid role argument type.
	ErrorInvalidRoleArgument = "invalid role argument type"
	// ErrorInvalidAccountArgument is the error message for invalid account argument.
	ErrorInvalidAccountArgument = "invalid account argument"
	// ErrorInvalidAmount is the error message for invalid amount.
	ErrorInvalidAmount = "invalid amount"
	// ErrorBurnAmountNotGreaterThanZero is the error message for burn amount not greater than 0.
	ErrorBurnAmountNotGreaterThanZero = "burn amount not greater than 0"
	// ErrorMintAmountNotGreaterThanZero is the error message for mint amount not greater than 0.
	ErrorMintAmountNotGreaterThanZero = "mint amount not greater than 0"
	// ErrorMintToZeroAddress is the error message for mint to the zero address.
	ErrorMintToZeroAddress = "mint to the zero address"
	// ErrorSenderNoRole is the error message when sender does not have the role.
	ErrorSenderNoRole = "access_control: sender does not have the role"
	// ErrorRenounceRoleDifferentThanCaller is the error message when renouncing role different from the caller.
	ErrorRenounceRoleDifferentThanCaller = "access_control: can only renounce roles for self"
)
