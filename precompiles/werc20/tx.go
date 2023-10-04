// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package werc20

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"
)

const (
	// DepositMethod defines the ABI method name for the WIERC20 deposit
	// transaction.
	DepositMethod = "deposit"
	// WithdrawMethod defines the ABI method name for the WIERC20 withdraw
	// transaction.
	WithdrawMethod = "withdraw"
)

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

	return nil, nil
}

func (p Precompile) Withdraw(
	ctx sdk.Context,
	contract *vm.Contract,
	stateDB vm.StateDB,
	_ *abi.Method,
	_ []interface{},
) ([]byte, error) {
	src := contract.Caller()
	amount := contract.Value()

	if err := p.EmitWithdrawEvent(ctx, stateDB, src, amount); err != nil {
		return nil, err
	}

	return nil, nil
}
