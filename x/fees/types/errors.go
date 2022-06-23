package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// errors
var (
	ErrFeesDisabled = sdkerrors.Register(ModuleName, 2, "developer fees disabled by governance")
	ErrFeeNotFound  = sdkerrors.Register(ModuleName, 3, "dev fee info not found")
	ErrFeeExists    = sdkerrors.Register(ModuleName, 4, "dev fee info for contract already registered")
)
