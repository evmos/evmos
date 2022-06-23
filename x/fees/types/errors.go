package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// errors
var (
	ErrInternalFee               = sdkerrors.Register(ModuleName, 2, "internal fees error")
	ErrFeesDisabled              = sdkerrors.Register(ModuleName, 3, "module is disabled")
	ErrFeesAlreadyRegistered     = sdkerrors.Register(ModuleName, 4, "contract fee already exists")
	ErrFeesNoContractDeployed    = sdkerrors.Register(ModuleName, 5, "no contract deployed")
	ErrFeesContractNotRegistered = sdkerrors.Register(ModuleName, 6, "no Fee registered for contract")
	ErrFeesDeployerIsNotEOA      = sdkerrors.Register(ModuleName, 7, "no Fee registered for contract")
)
