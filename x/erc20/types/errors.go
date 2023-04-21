// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package types

import (
	errorsmod "cosmossdk.io/errors"
)

// errors
var (
	ErrERC20Disabled          = errorsmod.Register(ModuleName, 2, "erc20 module is disabled")
	ErrInternalTokenPair      = errorsmod.Register(ModuleName, 3, "internal ethereum token mapping error")
	ErrTokenPairNotFound      = errorsmod.Register(ModuleName, 4, "token pair not found")
	ErrTokenPairAlreadyExists = errorsmod.Register(ModuleName, 5, "token pair already exists")
	ErrUndefinedOwner         = errorsmod.Register(ModuleName, 6, "undefined owner of contract pair")
	ErrBalanceInvariance      = errorsmod.Register(ModuleName, 7, "post transfer balance invariant failed")
	ErrUnexpectedEvent        = errorsmod.Register(ModuleName, 8, "unexpected event")
	ErrABIPack                = errorsmod.Register(ModuleName, 9, "contract ABI pack failed")
	ErrABIUnpack              = errorsmod.Register(ModuleName, 10, "contract ABI unpack failed")
	ErrEVMDenom               = errorsmod.Register(ModuleName, 11, "EVM denomination registration")
	ErrEVMCall                = errorsmod.Register(ModuleName, 12, "EVM call unexpected error")
	ErrERC20TokenPairDisabled = errorsmod.Register(ModuleName, 13, "erc20 token pair is disabled")
)
