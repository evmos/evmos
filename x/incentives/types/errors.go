package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// errors
var (
	ErrInternalIncentive = sdkerrors.Register(ModuleName, 2, "internal incentives error")
)
