package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// RootCodespace is the codespace for all errors defined in this package
const RootCodespace = "evoblock"

// root error codes for Evoblock
const (
	codeKeyTypeNotSupported = iota + 2
)

// errors
var (
	ErrKeyTypeNotSupported = sdkerrors.Register(RootCodespace, codeKeyTypeNotSupported, "key type 'secp256k1' not supported")
)
