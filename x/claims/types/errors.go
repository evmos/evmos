package types

import (
	sdkerrors "cosmossdk.io/errors"
)

// errors
var (
	ErrClaimsRecordNotFound = sdkerrors.Register(ModuleName, 2, "claims record not found")
	ErrInvalidAction        = sdkerrors.Register(ModuleName, 3, "invalid claim action type")
)
