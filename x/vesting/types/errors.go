package types

import (
	errorsmod "cosmossdk.io/errors"
)

// errors
var (
	ErrInsufficientVestedCoins = errorsmod.Register(ModuleName, 2, "insufficient vested coins error")
	ErrVestingLockup           = errorsmod.Register(ModuleName, 3, "vesting lockup error")
)
