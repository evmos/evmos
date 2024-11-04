// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	vestingtypes "github.com/evmos/evmos/v20/x/vesting/types"
)

// EthVestingExpenseTracker tracks both the total transaction value to be sent across Ethereum
// messages and the maximum spendable value for a given account.
type EthVestingExpenseTracker struct {
	// Total is the total value to be spent across a transaction with one or
	// more Ethereum message calls.
	Total *big.Int
	// Spendable is the maximum value that can be spent
	Spendable *big.Int
}

// CheckVesting checks if the account is a clawback vesting account and if so,
// checks that the account has sufficient unlocked balances to cover the
// transaction.
// The field `newExpenses` is expected to be in a 18 decimals representation.
func CheckVesting(
	ctx sdk.Context,
	evmKeeper EVMKeeper,
	account sdk.AccountI,
	accountExpenses map[string]*EthVestingExpenseTracker,
	newExpenses *big.Int,
) error {
	clawbackAccount, isClawback := account.(*vestingtypes.ClawbackVestingAccount)
	if !isClawback {
		return nil
	}

	// Check to make sure that the account does not exceed its spendable balances.
	// This transaction would fail in processing, so we should prevent it from
	// moving past the AnteHandler.
	expenses, err := UpdateAccountExpenses(ctx, evmKeeper, accountExpenses, clawbackAccount, newExpenses)
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

// UpdateAccountExpenses updates or sets the expenses for the given account, then
// returns the new value.
func UpdateAccountExpenses(
	ctx sdk.Context,
	evmKeeper EVMKeeper,
	accountExpenses map[string]*EthVestingExpenseTracker,
	account *vestingtypes.ClawbackVestingAccount,
	newExpenses *big.Int,
) (*EthVestingExpenseTracker, error) {
	address := account.GetAddress()
	addrStr := address.String()

	expenses, ok := accountExpenses[addrStr]
	// if an expense tracker is found for the address, add the expense and return
	if ok {
		expenses.Total = expenses.Total.Add(expenses.Total, newExpenses)
		return expenses, nil
	}

	// Get the account balance via EVM Keeper to be sure to have a 18 decimals
	// representation of the address balance.
	baseAmount := evmKeeper.GetBalance(ctx, common.BytesToAddress(address.Bytes()))

	// Short-circuit if the amount of base token is zero, since we require a non-zero balance to cover
	// gas fees at a minimum (these are defined to be non-zero). Note that this check
	// should be removed if the BaseFee definition is changed such that it can be zero.
	if baseAmount.Sign() == 0 {
		return nil, errorsmod.Wrapf(errortypes.ErrInsufficientFunds,
			"account has no balance to execute transaction: %s", addrStr)
	}

	// Get only base denom locked coins.
	lockedBalances := account.LockedCoins(ctx.BlockTime())
	lockedBaseAmount := math.ZeroInt()
	ok, lockedBaseBalance := lockedBalances.Find(evmtypes.GetEVMCoinDenom())
	if ok {
		lockedBaseAmount = lockedBaseBalance.Amount
	}

	lockedAmount := evmtypes.ConvertAmountTo18DecimalsBigInt(lockedBaseAmount.BigInt())

	spendableValue := big.NewInt(0)
	if amountDiff := new(big.Int).Sub(baseAmount, lockedAmount); amountDiff.Sign() > 0 {
		spendableValue = amountDiff
	}

	expenses = &EthVestingExpenseTracker{
		Total:     newExpenses,
		Spendable: spendableValue,
	}

	accountExpenses[addrStr] = expenses

	return expenses, nil
}
