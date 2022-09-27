package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// errors
var (
	ErrInternalRevenue              = sdkerrors.Register(ModuleName, 2, "internal revenue error")
	ErrRevenueDisabled              = sdkerrors.Register(ModuleName, 3, "revenue module is disabled by governance")
	ErrRevenueAlreadyRegistered     = sdkerrors.Register(ModuleName, 4, "revenue already exists for given contract")
	ErrRevenueNoContractDeployed    = sdkerrors.Register(ModuleName, 5, "no contract deployed")
	ErrRevenueContractNotRegistered = sdkerrors.Register(ModuleName, 6, "no revenue registered for contract")
	ErrRevenueDeployerIsNotEOA      = sdkerrors.Register(ModuleName, 7, "no revenue registered for contract")
)
