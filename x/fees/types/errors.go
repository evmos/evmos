package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// errors
var (
	ErrContractAlreadyRegistered    = sdkerrors.Register(ModuleName, 2, "contract already registered for fee distribution rewards")
	ErrContractWithdrawAddrNotFound = sdkerrors.Register(ModuleName, 3, "contract withdraw address not found")
)
