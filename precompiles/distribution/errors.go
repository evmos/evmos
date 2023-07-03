// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package distribution

const (
	// ErrSetWithdrawAddrAuth is raised when no authorization to set the withdraw address exists.
	ErrSetWithdrawAddrAuth = "set withdrawer address authorization for address %s does not exist"
	// ErrWithdrawDelRewardsAuth is raised when no authorization to withdraw delegation rewards exists.
	ErrWithdrawDelRewardsAuth = "withdraw delegation rewards authorization for address %s does not exist"
	// ErrWithdrawValCommissionAuth is raised when no authorization to withdraw validator commission exists.
	ErrWithdrawValCommissionAuth = "withdraw validator commission authorization for address %s does not exist"
	// ErrDifferentValidator is raised when the origin address is not the same as the validator address.
	ErrDifferentValidator = "origin address %s is not the same as validator address %s"
)
