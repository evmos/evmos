package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// errors
var (
	ErrClaimRecordNotFound = sdkerrors.Register(ModuleName, 2, "claim record not found")
)
