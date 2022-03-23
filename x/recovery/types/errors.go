package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// errors
var (
	ErrBlockedAddress = sdkerrors.Register(ModuleName, 2, "blocked address")
)
