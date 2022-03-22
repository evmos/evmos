package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// errors
var (
	ErrClaimsRecordNotFound = sdkerrors.Register(ModuleName, 2, "claims record not found")
	ErrInvalidAction        = sdkerrors.Register(ModuleName, 3, "invalid claim action type")
)
