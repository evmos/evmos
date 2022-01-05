package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// errors
var (
	ErrInvalidErc20Address      = sdkerrors.Register(ModuleName, 2, "invalid erc20 address")
	ErrUnmatchingCosmosDenom    = sdkerrors.Register(ModuleName, 3, "unmatching cosmos denom")
	ErrNotAllowedBridge         = sdkerrors.Register(ModuleName, 4, "not allowed bridge")
	ErrInternalEthMinting       = sdkerrors.Register(ModuleName, 5, "internal ethereum minting error")
	ErrWritingEthTxPayload      = sdkerrors.Register(ModuleName, 6, "writing ethereum tx payload error")
	ErrInternalTokenPair        = sdkerrors.Register(ModuleName, 7, "internal ethereum token mapping error")
	ErrUndefinedOwner           = sdkerrors.Register(ModuleName, 8, "undefined owner of contract pair")
	ErrSuicidedContract         = sdkerrors.Register(ModuleName, 9, "suicided contract pair")
	ErrInvalidConversionBalance = sdkerrors.Register(ModuleName, 10, "invalid conversion balance")
)
