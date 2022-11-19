package types

import (
	errorsmod "cosmossdk.io/errors"
)

// errors
var (
	ErrBlockedAddress = errorsmod.Register(ModuleName, 2, "blocked address")
)
