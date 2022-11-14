package types

import (
	errorsmod "cosmossdk.io/errors"
)

// errors
var (
	ErrInternalIncentive = errorsmod.Register(ModuleName, 2, "internal incentives error")
)
