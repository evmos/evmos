package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/unigov module sentinel errors
var (
	ErrUniGov = sdkerrors.Register(ModuleName, 1100, "unigov error")
	
)
