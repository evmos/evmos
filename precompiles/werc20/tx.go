// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package werc20

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
)

const (
	// DepositMethod defines the ABI method name for the IWERC20 deposit
	// transaction.
	DepositMethod = "deposit"
	// WithdrawMethod defines the ABI method name for the IWERC20 withdraw
	// transaction.
	WithdrawMethod = "withdraw"
)

// Deposit handles the payable deposit function. It retrieves the deposited amount
// and sends it back to the sender using the bank keeper.
func (p Precompile) Deposit(ctx sdk.Context, contract *vm.Contract, stateDB vm.StateDB) ([]byte, error) {
	caller := contract.Caller()
	depositedAmount := contract.Value()

	callerAccAddress := sdk.AccAddress(caller.Bytes())
	precompileAccAddr := sdk.AccAddress(p.Address().Bytes())

	// Send the coins back to the sender
	if err := p.BankKeeper.SendCoins(
		ctx,
		precompileAccAddr,
		callerAccAddress,
		sdk.NewCoins(sdk.Coin{
			Denom:  evmtypes.GetEVMCoinDenom(),
			Amount: math.NewIntFromBigInt(depositedAmount),
		}),
	); err != nil {
		return nil, err
	}

	// Add the entries to the statedb journal in 18 decimals.
	convertedAmount := evmtypes.ConvertAmountTo18DecimalsBigInt(depositedAmount)
	p.SetBalanceChangeEntries(
		cmn.NewBalanceChangeEntry(caller, convertedAmount, cmn.Add),
		cmn.NewBalanceChangeEntry(p.Address(), convertedAmount, cmn.Sub),
	)

	if err := p.EmitDepositEvent(ctx, stateDB, caller, depositedAmount); err != nil {
		return nil, err
	}

	return nil, nil
}

// Withdraw is a no-op and mock function that provides the same interface as the
// WETH contract to support equality between the native coin and its wrapped
// ERC-20 (e.g. EVMOS and WEVMOS).
func (p Precompile) Withdraw(ctx sdk.Context, contract *vm.Contract, stateDB vm.StateDB, args []interface{}) ([]byte, error) {
	amount, ok := args[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid argument type: %T", args[0])
	}
	// TODO: check that the sender has enough balance for the withdraw.
	if err := p.EmitWithdrawalEvent(ctx, stateDB, contract.Caller(), amount); err != nil {
		return nil, err
	}
	return nil, nil
}
