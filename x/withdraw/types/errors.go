package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// errors
var (
	ErrKeyTypeNotSupported = sdkerrors.Register(ModuleName, 2, "key type 'secp256k1' not supported")
)
