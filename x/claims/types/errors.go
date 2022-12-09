package types

import (
	errorsmod "cosmossdk.io/errors"
)

// errors
var (
	ErrClaimsRecordNotFound = errorsmod.Register(ModuleName, 2, "claims record not found")
	ErrInvalidAction        = errorsmod.Register(ModuleName, 3, "invalid claim action type")
)
