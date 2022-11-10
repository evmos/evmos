package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// IBC EMV transaction sentinel errors
var (
	ErrInvalidPacketTimeout    = sdkerrors.Register(ModuleName, 2, "invalid packet timeout")
	ErrInvalidVersion          = sdkerrors.Register(ModuleName, 4, "invalid evm-tx version")
	ErrSendDisabled            = sdkerrors.Register(ModuleName, 7, "evm transactions from this chain are disabled")
	ErrReceiveDisabled         = sdkerrors.Register(ModuleName, 8, "evm transactions to this chain are disabled")
	ErrMaxTransferChannels     = sdkerrors.Register(ModuleName, 9, "max transfer channels")
)
