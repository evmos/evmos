// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package authorization

const (
	// ErrAuthzDoesNotExistOrExpired is raised when the authorization does not exist.
	ErrAuthzDoesNotExistOrExpired = "authorization to %s for address %s does not exist or is expired"
	// ErrEmptyMethods is raised when the given methods array is empty.
	ErrEmptyMethods = "no methods defined; expected at least one message type url"
	// ErrEmptyStringInMethods is raised when the given methods array contains an empty string.
	ErrEmptyStringInMethods = "empty string found in methods array; expected no empty strings to be passed; got: %v"
	// ErrExceededAllowance is raised when the amount exceeds the set allowance.
	ErrExceededAllowance = "amount %s greater than allowed limit %s"
	// ErrInvalidGranter is raised when the granter address is not valid.
	ErrInvalidGranter = "invalid granter address: %v"
	// ErrInvalidGrantee is raised when the grantee address is not valid.
	ErrInvalidGrantee = "invalid grantee address: %v"
	// ErrInvalidMethods is raised when the given methods cannot be unpacked.
	ErrInvalidMethods = "invalid methods defined; expected an array of strings; got: %v"
	// ErrInvalidMethod is raised when the given method cannot be unpacked.
	ErrInvalidMethod = "invalid method defined; expected a string; got: %v"
	// ErrAuthzNotAccepted is raised when the authorization is not accepted.
	ErrAuthzNotAccepted = "authorization to %s for address %s is not accepted"
)
