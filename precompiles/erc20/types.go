// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package erc20

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// NewMsgSend creates a new MsgSend instance and does sanity checks
// on the given arguments before populating the message.
func ParseTransferArgs(args []interface{}) (
	to common.Address, amount *big.Int, err error,
) {
	if len(args) != 2 {
		return common.Address{}, nil, fmt.Errorf("invalid number of arguments; expected 2; got: %d", len(args))
	}

	to, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, nil, fmt.Errorf("invalid to address: %v", args[0])
	}

	amount, ok = args[1].(*big.Int)
	if !ok {
		return common.Address{}, nil, fmt.Errorf("invalid amount")
	}

	return to, amount, nil
}

// NewMsgSendFrom creates a new MsgSend instance and does sanity checks
// on the given arguments before populating the message.
func ParseTransferFromArgs(args []interface{}) (
	from common.Address, to common.Address, amount *big.Int, err error,
) {
	if len(args) != 3 {
		return common.Address{}, common.Address{}, nil, fmt.Errorf("invalid number of arguments; expected 3; got: %d", len(args))
	}

	from, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, nil, fmt.Errorf("invalid from address: %v", args[0])
	}

	to, ok = args[1].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, nil, fmt.Errorf("invalid to address: %v", args[0])
	}

	amount, ok = args[2].(*big.Int)
	if !ok {
		return common.Address{}, common.Address{}, nil, fmt.Errorf("invalid amount")
	}

	return from, to, amount, nil
}

// ParseApproveArgs parses the arguments and returns the spender and amount
func ParseApproveArgs(args []interface{}) (
	spender common.Address, amount *big.Int, err error,
) {
	if len(args) != 2 {
		return common.Address{}, nil, fmt.Errorf("invalid number of arguments; expected 2; got: %d", len(args))
	}

	spender, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, nil, fmt.Errorf("invalid spender address: %v", args[0])
	}

	amount, ok = args[1].(*big.Int)
	if !ok {
		return common.Address{}, nil, fmt.Errorf("invalid amount")
	}

	return spender, amount, nil
}

// ParseAllowanceArgs parses the arguments and returns the owner and spender
func ParseAllowanceArgs(args []interface{}) (
	owner, spender common.Address, err error,
) {
	if len(args) != 2 {
		return common.Address{}, common.Address{}, fmt.Errorf("invalid number of arguments; expected 2; got: %d", len(args))
	}

	owner, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, fmt.Errorf("invalid owner address: %v", args[0])
	}

	spender, ok = args[1].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, fmt.Errorf("invalid spender address: %v", args[0])
	}

	return owner, spender, nil
}
