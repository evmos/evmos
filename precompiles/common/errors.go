// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package common

const (
	// ErrNotRunInEvm is raised when a function is not called inside the EVM.
	ErrNotRunInEvm = "not run in EVM"
	// ErrDifferentOrigin is raised when an approval is set but the origin address is not the same as the spender.
	ErrDifferentOrigin = "tx origin address %s does not match the delegator address %s"
	// ErrInvalidABI is raised when the ABI cannot be parsed.
	ErrInvalidABI = "invalid ABI: %w"
	// ErrInvalidAmount is raised when the amount cannot be cast to a big.Int.
	ErrInvalidAmount = "invalid amount: %v"
	// ErrInvalidDelegator is raised when the delegator address is not valid.
	ErrInvalidDelegator = "invalid delegator address: %s"
	// ErrInvalidDenom is raised when the denom is not valid.
	ErrInvalidDenom = "invalid denom: %s"
	// ErrInvalidMsgType is raised when the transaction type is not valid for the given precompile.
	ErrInvalidMsgType = "invalid %s transaction type: %s"
	// ErrInvalidNumberOfArgs is raised when the number of arguments is not what is expected.
	ErrInvalidNumberOfArgs = "invalid number of arguments; expected %d; got: %d"
	// ErrUnknownMethod is raised when the method is not known.
	ErrUnknownMethod = "unknown method: %s"
	// ErrIntegerOverflow is raised when an integer overflow occurs.
	ErrIntegerOverflow = "integer overflow when increasing allowance"
	// ErrNegativeAmount is raised when an amount is negative.
	ErrNegativeAmount = "negative amount when decreasing allowance"
	// ErrInvalidType is raised when the provided type is different than the expected.
	ErrInvalidType = "invalid type for %s: expected %T, received %T"
)
