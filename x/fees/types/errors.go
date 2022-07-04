package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// errors
var (
	ErrInternalFee               = sdkerrors.Register(ModuleName, 2, "internal fees error")
	ErrFeesDisabled              = sdkerrors.Register(ModuleName, 3, "fees module is disabled by governance")
	ErrFeesAlreadyRegistered     = sdkerrors.Register(ModuleName, 4, "fee already exists for given contract")
	ErrFeesNoContractDeployed    = sdkerrors.Register(ModuleName, 5, "no contract deployed")
	ErrFeesContractNotRegistered = sdkerrors.Register(ModuleName, 6, "no fee registered for contract")
	ErrFeesDeployerIsNotEOA      = sdkerrors.Register(ModuleName, 7, "no fee registered for contract")
)
