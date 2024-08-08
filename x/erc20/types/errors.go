// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	errorsmod "cosmossdk.io/errors"
)

// errors
var (
	ErrERC20Disabled            = errorsmod.Register(ModuleName, 2, "erc20 module is disabled")
	ErrInternalTokenPair        = errorsmod.Register(ModuleName, 3, "internal ethereum token mapping error")
	ErrTokenPairNotFound        = errorsmod.Register(ModuleName, 4, "token pair not found")
	ErrTokenPairAlreadyExists   = errorsmod.Register(ModuleName, 5, "token pair already exists")
	ErrUndefinedOwner           = errorsmod.Register(ModuleName, 6, "undefined owner of contract pair")
	ErrBalanceInvariance        = errorsmod.Register(ModuleName, 7, "post transfer balance invariant failed")
	ErrUnexpectedEvent          = errorsmod.Register(ModuleName, 8, "unexpected event")
	ErrABIPack                  = errorsmod.Register(ModuleName, 9, "contract ABI pack failed")
	ErrABIUnpack                = errorsmod.Register(ModuleName, 10, "contract ABI unpack failed")
	ErrEVMDenom                 = errorsmod.Register(ModuleName, 11, "EVM denomination registration")
	ErrEVMCall                  = errorsmod.Register(ModuleName, 12, "EVM call unexpected error")
	ErrERC20TokenPairDisabled   = errorsmod.Register(ModuleName, 13, "erc20 token pair is disabled")
	ErrInvalidIBC               = errorsmod.Register(ModuleName, 14, "invalid IBC transaction")
	ErrTokenPairOwnedByModule   = errorsmod.Register(ModuleName, 15, "token pair owned by module")
	ErrNativeConversionDisabled = errorsmod.Register(ModuleName, 16, "native coins manual conversion is disabled")
)
