// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
	vestingtypes "github.com/evmos/evmos/v19/x/vesting/types"
)

// EthVestingTransactionDecorator validates if clawback vesting accounts are
// permitted to perform Ethereum Tx.
type EthVestingTransactionDecorator struct {
	ak evmtypes.AccountKeeper
	bk evmtypes.BankKeeper
	ek EVMKeeper
}

// EthVestingExpenseTracker tracks both the total transaction value to be sent across Ethereum
// messages and the maximum spendable value for a given account.
type EthVestingExpenseTracker struct {
	// Total is the total value to be spent across a transaction with one or more Ethereum message calls
	Total *big.Int
	// Spendable is the maximum value that can be spent
	Spendable *big.Int
}

// NewEthVestingTransactionDecorator returns a new EthVestingTransactionDecorator.
//
// NOTE: Can't delete the legacy decorator yet because vesting module's tests would have to be refactored
func NewEthVestingTransactionDecorator(ak evmtypes.AccountKeeper, bk evmtypes.BankKeeper, ek EVMKeeper) EthVestingTransactionDecorator {
	return EthVestingTransactionDecorator{
		ak: ak,
		bk: bk,
		ek: ek,
	}
}

// AnteHandle validates that a clawback vesting account has surpassed the
// vesting cliff and lockup period.
//
// This AnteHandler decorator will fail if:
//   - the message is not a MsgEthereumTx
//   - sender account cannot be found
//   - tx values are in excess of any account's spendable balances
func (vtd EthVestingTransactionDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// Track the total value to be spent by each address across all messages and ensure
	// that no account can exceed its spendable balance.
	accountExpenses := make(map[string]*EthVestingExpenseTracker)
	denom := vtd.ek.GetParams(ctx).EvmDenom

	for _, msg := range tx.GetMsgs() {
		_, txData, from, err := evmtypes.UnpackEthMsg(msg)
		if err != nil {
			return ctx, err
		}

		value := txData.GetValue()

		acc := vtd.ak.GetAccount(ctx, from)
		if acc == nil {
			return ctx, errorsmod.Wrapf(errortypes.ErrUnknownAddress,
				"account %s does not exist", acc)
		}

		if err := CheckVesting(ctx, vtd.bk, acc, accountExpenses, value, denom); err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

// CheckVesting checks if the account is a clawback vesting account and if so,
// checks that the account has sufficient unlocked balances to cover the
// transaction.
func CheckVesting(
	ctx sdk.Context,
	bankKeeper evmtypes.BankKeeper,
	account authtypes.AccountI,
	accountExpenses map[string]*EthVestingExpenseTracker,
	addedExpense *big.Int,
	denom string,
) error {
	clawbackAccount, isClawback := account.(*vestingtypes.ClawbackVestingAccount)
	if !isClawback {
		return nil
	}

	// Check to make sure that the account does not exceed its spendable balances.
	// This transaction would fail in processing, so we should prevent it from
	// moving past the AnteHandler.

	expenses, err := UpdateAccountExpenses(ctx, bankKeeper, accountExpenses, clawbackAccount, addedExpense, denom)
	if err != nil {
		return err
	}

	total := expenses.Total
	spendable := expenses.Spendable

	if total.Cmp(spendable) > 0 {
		return errorsmod.Wrapf(vestingtypes.ErrInsufficientUnlockedCoins,
			"clawback vesting account has insufficient unlocked tokens to execute transaction: %s < %s", spendable.String(), total.String(),
		)
	}

	return nil
}

// UpdateAccountExpenses updates or sets the totalSpend for the given account, then
// returns the new value.
func UpdateAccountExpenses(
	ctx sdk.Context,
	bankKeeper evmtypes.BankKeeper,
	accountExpenses map[string]*EthVestingExpenseTracker,
	account *vestingtypes.ClawbackVestingAccount,
	addedExpense *big.Int,
	denom string,
) (*EthVestingExpenseTracker, error) {
	address := account.GetAddress()
	addrStr := address.String()

	expenses, ok := accountExpenses[addrStr]
	// if an expense tracker is found for the address, add the expense and return
	if ok {
		expenses.Total = expenses.Total.Add(expenses.Total, addedExpense)
		return expenses, nil
	}

	balance := bankKeeper.GetBalance(ctx, address, denom)

	// Short-circuit if the balance is zero, since we require a non-zero balance to cover
	// gas fees at a minimum (these are defined to be non-zero). Note that this check
	// should be removed if the BaseFee definition is changed such that it can be zero.
	if balance.IsZero() {
		return nil, errorsmod.Wrapf(errortypes.ErrInsufficientFunds,
			"account has no balance to execute transaction: %s", addrStr)
	}

	lockedBalances := account.LockedCoins(ctx.BlockTime())
	ok, lockedBalance := lockedBalances.Find(denom)
	if !ok {
		lockedBalance = sdk.Coin{Denom: denom, Amount: math.ZeroInt()}
	}

	spendableValue := big.NewInt(0)
	if spendableBalance, err := balance.SafeSub(lockedBalance); err == nil {
		spendableValue = spendableBalance.Amount.BigInt()
	}

	expenses = &EthVestingExpenseTracker{
		Total:     addedExpense,
		Spendable: spendableValue,
	}

	accountExpenses[addrStr] = expenses

	return expenses, nil
}
