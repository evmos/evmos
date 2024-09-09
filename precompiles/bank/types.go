// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package bank

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
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
		return nil, fmt.Errorf(cmn.ErrInvalidType, "account", common.Address{}, args[0])
	}

	return account.Bytes(), nil
}

// ParseSupplyOfArgs parses the call arguments for the bank SupplyOf query.
func ParseSupplyOfArgs(args []interface{}) (common.Address, error) {
	if len(args) != 1 {
		return common.Address{}, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	erc20Address, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, fmt.Errorf(cmn.ErrInvalidType, "erc20Address", common.Address{}, args[0])
	}

	return erc20Address, nil
}
