// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	errorsmod "cosmossdk.io/errors"
)

// errors
var (
	ErrInsufficientVestedCoins   = errorsmod.Register(ModuleName, 2, "insufficient vested coins error")
	ErrVestingLockup             = errorsmod.Register(ModuleName, 3, "vesting lockup error")
	ErrInsufficientUnlockedCoins = errorsmod.Register(ModuleName, 4, "insufficient unlocked coins error")
	ErrNothingToClawback         = errorsmod.Register(ModuleName, 5, "nothing to clawback from the account")
	ErrNotSubjectToClawback      = errorsmod.Register(ModuleName, 6, "account is not subject to clawback vesting")
	ErrNotSubjectToGovClawback   = errorsmod.Register(ModuleName, 7, "account does not have governance clawback enabled")
)
