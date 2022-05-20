package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// errors
var (
	ErrInternalFee = sdkerrors.Register(ModuleName, 2, "internal fees error")
)
