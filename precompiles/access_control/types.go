// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package accesscontrol

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// ParseRoleArgs parses the role arguments.
func ParseRoleArgs(args []interface{}) (common.Hash, common.Address, error) {
	if len(args) != 2 {
		return common.Hash{}, common.Address{}, fmt.Errorf(ErrorInvalidArgumentNumber)
	}
	roleArray, ok := args[0].([32]uint8)
	if !ok {
		return common.Hash{}, common.Address{}, fmt.Errorf(ErrorInvalidRoleArgument)
	}

	var role common.Hash
	copy(role[:], roleArray[:])

	account, ok := args[1].(common.Address)
	if !ok {
		return common.Hash{}, common.Address{}, fmt.Errorf(ErrorInvalidAccountArgument)
	}

	return role, account, nil
}

// ParseBurnArgs parses the burn arguments.
func ParseBurnArgs(args []interface{}) (*big.Int, error) {
	amount, ok := args[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf(ErrorInvalidAmount)
	}

	if amount.Sign() != 1 {
		return nil, fmt.Errorf(ErrorBurnAmountNotGreaterThanZero)
	}

	return amount, nil
}

// ParseMintArgs parses the mint arguments.
func ParseMintArgs(args []interface{}) (common.Address, *big.Int, error) {
	to, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, nil, fmt.Errorf(ErrorInvalidMinterAddress)
	}

	if to == (common.Address{}) {
		return common.Address{}, nil, fmt.Errorf(ErrorMintToZeroAddress)
	}

	amount, ok := args[1].(*big.Int)
	if !ok {
		return common.Address{}, nil, fmt.Errorf(ErrorInvalidAmount)
	}

	if amount.Sign() != 1 {
		return common.Address{}, nil, fmt.Errorf(ErrorMintAmountNotGreaterThanZero)
	}

	return to, amount, nil
}
