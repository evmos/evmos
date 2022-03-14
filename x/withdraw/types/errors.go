package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// errors
var (
	ErrChannelNotEnabled = sdkerrors.Register(ModuleName, 2, "channel not enabled")
)
