// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package erc20

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
)

// EventTransfer defines the event data for the ERC20 Transfer events.
type EventTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
}

// EventApproval defines the event data for the ERC20 Approval events.
type EventApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
}

// ParseTransferArgs parses the arguments from the transfer method and returns
// the destination address (to) and amount.
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
		return common.Address{}, nil, fmt.Errorf("invalid amount: %v", args[1])
	}

	return to, amount, nil
}

// ParseTransferFromArgs parses the arguments from the transferFrom method and returns
// the sender address (from), destination address (to) and amount.
func ParseTransferFromArgs(args []interface{}) (
	from, to common.Address, amount *big.Int, err error,
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
		return common.Address{}, common.Address{}, nil, fmt.Errorf("invalid to address: %v", args[1])
	}

	amount, ok = args[2].(*big.Int)
	if !ok {
		return common.Address{}, common.Address{}, nil, fmt.Errorf("invalid amount: %v", args[2])
	}

	return from, to, amount, nil
}

// ParseApproveArgs parses the approval arguments and returns the spender address
// and amount.
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
		return common.Address{}, nil, fmt.Errorf("invalid amount: %v", args[1])
	}

	return spender, amount, nil
}

// ParseAllowanceArgs parses the allowance arguments and returns the owner and
// the spender addresses.
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
		return common.Address{}, common.Address{}, fmt.Errorf("invalid spender address: %v", args[1])
	}

	return owner, spender, nil
}

// ParseBalanceOfArgs parses the balanceOf arguments and returns the account address.
func ParseBalanceOfArgs(args []interface{}) (common.Address, error) {
	if len(args) != 1 {
		return common.Address{}, fmt.Errorf("invalid number of arguments; expected 1; got: %d", len(args))
	}

	account, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, fmt.Errorf("invalid account address: %v", args[0])
	}

	return account, nil
}

// updateOrAddCoin replaces the coin of the given denomination in the coins slice or adds it if it
// does not exist yet.
//
// CONTRACT: Requires the coins struct to contain at most one coin of the given
// denom.
func updateOrAddCoin(coins sdk.Coins, coin sdk.Coin) sdk.Coins {
	for idx, c := range coins {
		if c.Denom == coin.Denom {
			coins[idx] = coin
			return coins
		}
	}

	// NOTE: if no coin with the correct denomination is in the coins slice, we
	// add it here.
	return coins.Add(coin)
}
