package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// transfer hooks sentinel errors
var (
	ErrInvalidVersion = sdkerrors.Register(ModuleName, 2, "invalid ICS20 transfer hooks middleware version")
)
