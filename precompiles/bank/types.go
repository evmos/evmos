// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package bank

import (
	"errors"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
)

// Balance contains the amount for a corresponding ERC-20 contract address
type Balance struct {
	ContractAddress common.Address
	Amount          *big.Int
}

// ParseBalancesArgs parses the call arguments for the bank Balances query.
func ParseBalancesArgs(args []interface{}) (sdk.AccAddress, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	account, ok := args[0].(common.Address)
	if !ok {
		return nil, errors.New("invalid account address")
	}

	return account.Bytes(), nil
}
