// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package accesscontrol

import (
	"fmt"
	commonerr "github.com/evmos/evmos/v18/precompiles/common"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// ParseRoleArgs parses the role arguments.
func ParseRoleArgs(args []interface{}) (common.Hash, common.Address, error) {
	if len(args) != 2 {
		return common.Hash{}, common.Address{}, fmt.Errorf(commonerr.ErrInvalidNumberOfArgs, 2, len(args))
	}
	roleArray, ok := args[0].([32]uint8)
	if !ok {
		return common.Hash{}, common.Address{}, fmt.Errorf(ErrInvalidRoleArgument)
	}

	var role common.Hash
	copy(role[:], roleArray[:])

	account, ok := args[1].(common.Address)
	if !ok {
		return common.Hash{}, common.Address{}, fmt.Errorf(ErrInvalidAccountArgument)
	}

	return role, account, nil
}

// ParseBurnArgs parses the burn arguments.
func ParseBurnArgs(args []interface{}) (*big.Int, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(commonerr.ErrInvalidNumberOfArgs, 1, len(args))
	}

	amount, ok := args[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf(commonerr.ErrInvalidAmount, amount)
	}

	if amount.Sign() != 1 {
		return nil, fmt.Errorf(ErrBurnAmountNotGreaterThanZero)
	}

	return amount, nil
}

// ParseMintArgs parses the mint arguments.
func ParseMintArgs(args []interface{}) (common.Address, *big.Int, error) {
	if len(args) != 2 {
		return common.Address{}, nil, fmt.Errorf(commonerr.ErrInvalidNumberOfArgs, 2, len(args))
	}

	to, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, nil, fmt.Errorf(ErrInvalidMinterAddress)
	}

	if to == (common.Address{}) {
		return common.Address{}, nil, fmt.Errorf(ErrMintToZeroAddress)
	}

	amount, ok := args[1].(*big.Int)
	if !ok {
		return common.Address{}, nil, fmt.Errorf(commonerr.ErrInvalidAmount, amount)
	}

	if amount.Sign() != 1 {
		return common.Address{}, nil, fmt.Errorf(ErrMintAmountNotGreaterThanZero)
	}

	return to, amount, nil
}
