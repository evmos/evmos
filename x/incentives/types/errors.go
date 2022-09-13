package types

import (
	sdkerrors "cosmossdk.io/errors"
)

// errors
var (
	ErrInternalIncentive = sdkerrors.Register(ModuleName, 2, "internal incentives error")
)
