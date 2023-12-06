// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package werc20

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"
)

const (
	// DepositMethod defines the ABI method name for the IWERC20 deposit
	// transaction.
	DepositMethod = "deposit"
	// WithdrawMethod defines the ABI method name for the IWERC20 withdraw
	// transaction.
	WithdrawMethod = "withdraw"
)

// Deposit is a no-op and mock function that provides the same interface as the
// WETH contract to support equality between the native coin and its wrapped
// ERC-20 (eg. EVMOS and WEVMOS). It only emits the Deposit event.
func (p Precompile) Deposit(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	_ *abi.Method,
	_ []interface{},
) ([]byte, error) {
	dst := contract.Caller()
	amount := contract.Value()

	if err := p.EmitDepositEvent(ctx, stateDB, dst, amount); err != nil {
		return nil, err
	}

	// NOTE: To avoid triggering the minting/burning mechanism we override the
	// balances of the contract and the sender manually to perform a no-op so
	// the balances are kept in sync.
	stateDB.AddBalance(dst, amount)
	stateDB.SubBalance(contract.Address(), amount)

	return nil, nil
}

// Withdraw is a no-op and mock function that provides the same interface as the
// WETH contract to support equality between the native coin and its wrapped
// ERC-20 (eg. EVMOS and WEVMOS). It only emits the Withdraw event.
func (p Precompile) Withdraw(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	_ *abi.Method,
	args []interface{},
) ([]byte, error) {
	src := contract.Caller()
	amount, ok := args[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid argument type: %T", args[0])
	}

	if err := p.EmitWithdrawalEvent(ctx, stateDB, src, amount); err != nil {
		return nil, err
	}

	return nil, nil
}
