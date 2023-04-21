// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	evmtypes "github.com/evmos/evmos/v12/x/evm/types"
	vestingtypes "github.com/evmos/evmos/v12/x/vesting/types"
)

// EthVestingTransactionDecorator validates if clawback vesting accounts are
// permitted to perform Ethereum Tx.
type EthVestingTransactionDecorator struct {
	ak evmtypes.AccountKeeper
	bk evmtypes.BankKeeper
	ek EVMKeeper
}

// ethVestingExpenseTracker tracks both the total transaction value to be sent across Ethereum
// messages and the maximum spendable value for a given account.
type ethVestingExpenseTracker struct {
	// total is the total value to be spent across a transaction with one or more Ethereum message calls
	total *big.Int
	// spendable is the maximum value that can be spent
	spendable *big.Int
}

// NewEthVestingTransactionDecorator returns a new EthVestingTransactionDecorator.
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
	accountExpenses := make(map[string]*ethVestingExpenseTracker)
	denom := vtd.ek.GetParams(ctx).EvmDenom

	for _, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, errorsmod.Wrapf(errortypes.ErrUnknownRequest,
				"invalid message type %T, expected %T", msg, (*evmtypes.MsgEthereumTx)(nil),
			)
		}

		acc := vtd.ak.GetAccount(ctx, msgEthTx.GetFrom())
		if acc == nil {
			return ctx, errorsmod.Wrapf(errortypes.ErrUnknownAddress,
				"account %s does not exist", acc)
		}

		// Check that this decorator only applies to clawback vesting accounts
		clawbackAccount, isClawback := acc.(*vestingtypes.ClawbackVestingAccount)
		if !isClawback {
			continue
		}

		// Check to make sure that the account does not exceed its spendable balances.
		// This transaction would fail in processing, so we should prevent it from
		// moving past the AnteHandler.
		msgValue := msgEthTx.AsTransaction().Value()

		expenses, err := vtd.updateAccountExpenses(ctx, accountExpenses, clawbackAccount, msgValue, denom)
		if err != nil {
			return ctx, err
		}

		total := expenses.total
		spendable := expenses.spendable

		if total.Cmp(spendable) > 0 {
			return ctx, errorsmod.Wrapf(vestingtypes.ErrInsufficientUnlockedCoins,
				"clawback vesting account has insufficient unlocked tokens to execute transaction: %s < %s", spendable.String(), total.String(),
			)
		}
	}

	return next(ctx, tx, simulate)
}

// updateAccountExpenses updates or sets the totalSpend for the given account, then
// returns the new value.
func (vtd EthVestingTransactionDecorator) updateAccountExpenses(
	ctx sdk.Context,
	accountExpenses map[string]*ethVestingExpenseTracker,
	account *vestingtypes.ClawbackVestingAccount,
	addedExpense *big.Int,
	denom string,
) (*ethVestingExpenseTracker, error) {
	address := account.GetAddress()
	addrStr := address.String()

	expenses, ok := accountExpenses[addrStr]
	// if an expense tracker is found for the address, add the expense and return
	if ok {
		expenses.total = expenses.total.Add(expenses.total, addedExpense)
		return expenses, nil
	}

	balance := vtd.bk.GetBalance(ctx, address, denom)

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
		lockedBalance = sdk.Coin{Denom: denom, Amount: sdk.ZeroInt()}
	}

	spendableValue := big.NewInt(0)
	if spendableBalance, err := balance.SafeSub(lockedBalance); err == nil {
		spendableValue = spendableBalance.Amount.BigInt()
	}

	expenses = &ethVestingExpenseTracker{
		total:     addedExpense,
		spendable: spendableValue,
	}

	accountExpenses[addrStr] = expenses

	return expenses, nil
}
