// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package ante

import (
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ethante "github.com/evmos/ethermint/app/ante"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	vestingtypes "github.com/evmos/evmos/v11/x/vesting/types"
)

// EthVestingTransactionDecorator validates if clawback vesting accounts are
// permitted to perform Ethereum Tx.
type EthVestingTransactionDecorator struct {
	ak evmtypes.AccountKeeper
	bk evmtypes.BankKeeper
	ek ethante.EVMKeeper
}

// ethVestingTotalSpend tracks both the total transaction value to be sent across Ethereum
// messages and the maximum spendable value for a given account.
type ethVestingTotalSpend struct {
	// totalValue is the total value to be spent across a transaction with one or more Ethereum message calls
	totalValue *big.Int
	// spendableValue is the maximum value that can be spent
	spendableValue *big.Int
}

func NewEthVestingTransactionDecorator(ak evmtypes.AccountKeeper, bk evmtypes.BankKeeper, ek ethante.EVMKeeper) EthVestingTransactionDecorator {
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
	totalSpendByAddress := make(map[string]*ethVestingTotalSpend)
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

		totalSpend, err := vtd.getTotalSpend(ctx, totalSpendByAddress, clawbackAccount, msgValue, denom)
		if err != nil {
			return ctx, err
		}

		totalValue := totalSpend.totalValue
		spendableValue := totalSpend.spendableValue

		if totalValue.Cmp(spendableValue) > 0 {
			return ctx, errorsmod.Wrapf(vestingtypes.ErrInsufficientUnlockedCoins,
				"clawback vesting account has insufficient unlocked tokens to execute transaction: %s < %s", spendableValue.String(), totalValue.String(),
			)
		}
	}

	return next(ctx, tx, simulate)
}

// getTotalSpend updates or sets the totalSpend for the given account, then
// returns the new value.
func (vtd EthVestingTransactionDecorator) getTotalSpend(
	ctx sdk.Context,
	totalSpendByAddress map[string]*ethVestingTotalSpend,
	account *vestingtypes.ClawbackVestingAccount,
	value *big.Int,
	denom string,
) (*ethVestingTotalSpend, error) {
	address := account.GetAddress()

	totalSpend, ok := totalSpendByAddress[address.String()]
	if ok {
		totalSpend.totalValue.Add(totalSpend.totalValue, value)
		return totalSpend, nil
	}

	balance := vtd.bk.GetBalance(ctx, address, denom)

	// Short-circuit if the balance is zero, since we require a non-zero balance to cover
	// gas fees at a minimum (these are defined to be non-zero). Note that this check
	// should be removed if the BaseFee definition is changed such that it can be zero.
	if balance.IsZero() {
		return nil, errorsmod.Wrapf(errortypes.ErrInsufficientFunds,
			"account has no balance to execute transaction: %s", address.String())
	}

	ok, lockedBalance := account.LockedCoins(ctx.BlockTime()).Find(denom)
	if !ok {
		lockedBalance = sdk.NewCoin(denom, sdk.ZeroInt())
	}

	spendableValue := big.NewInt(0)
	if spendableBalance, err := balance.SafeSub(lockedBalance); err == nil {
		spendableValue = spendableBalance.Amount.BigInt()
	}

	totalSpend = &ethVestingTotalSpend{
		totalValue:     value,
		spendableValue: spendableValue,
	}

	totalSpendByAddress[address.String()] = totalSpend

	return totalSpend, nil
}

// TODO: remove once Cosmos SDK is upgraded to v0.46

// VestingDelegationDecorator validates delegation of vested coins
type VestingDelegationDecorator struct {
	ak  evmtypes.AccountKeeper
	sk  vestingtypes.StakingKeeper
	cdc codec.BinaryCodec
}

// NewVestingDelegationDecorator creates a new VestingDelegationDecorator
func NewVestingDelegationDecorator(ak evmtypes.AccountKeeper, sk vestingtypes.StakingKeeper, cdc codec.BinaryCodec) VestingDelegationDecorator {
	return VestingDelegationDecorator{
		ak:  ak,
		sk:  sk,
		cdc: cdc,
	}
}

// AnteHandle checks if the tx contains a staking delegation.
// It errors if the coins are still locked or the bond amount is greater than
// the coins already vested
func (vdd VestingDelegationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		switch msg := msg.(type) {
		case *authz.MsgExec:
			// Check for bypassing authorization
			if err := vdd.validateAuthz(ctx, msg); err != nil {
				return ctx, err
			}
		default:
			if err := vdd.validateMsg(ctx, msg); err != nil {
				return ctx, err
			}
		}
	}

	return next(ctx, tx, simulate)
}

// validateAuthz validates the authorization internal message
func (vdd VestingDelegationDecorator) validateAuthz(ctx sdk.Context, execMsg *authz.MsgExec) error {
	for _, v := range execMsg.Msgs {
		var innerMsg sdk.Msg
		if err := vdd.cdc.UnpackAny(v, &innerMsg); err != nil {
			return errorsmod.Wrap(err, "cannot unmarshal authz exec msgs")
		}

		if err := vdd.validateMsg(ctx, innerMsg); err != nil {
			return err
		}
	}

	return nil
}

// validateMsg checks that the only vested coins can be delegated
func (vdd VestingDelegationDecorator) validateMsg(ctx sdk.Context, msg sdk.Msg) error {
	delegateMsg, ok := msg.(*stakingtypes.MsgDelegate)
	if !ok {
		return nil
	}

	for _, addr := range msg.GetSigners() {
		acc := vdd.ak.GetAccount(ctx, addr)
		if acc == nil {
			return errorsmod.Wrapf(
				errortypes.ErrUnknownAddress,
				"account %s does not exist", addr,
			)
		}

		clawbackAccount, isClawback := acc.(*vestingtypes.ClawbackVestingAccount)
		if !isClawback {
			// continue to next decorator as this logic only applies to vesting
			return nil
		}

		// error if bond amount is > vested coins
		bondDenom := vdd.sk.BondDenom(ctx)
		coins := clawbackAccount.GetVestedOnly(ctx.BlockTime())
		if coins == nil || coins.Empty() {
			return errorsmod.Wrap(
				vestingtypes.ErrInsufficientVestedCoins,
				"account has no vested coins",
			)
		}

		vested := coins.AmountOf(bondDenom)
		if vested.LT(delegateMsg.Amount.Amount) {
			return errorsmod.Wrapf(
				vestingtypes.ErrInsufficientVestedCoins,
				"cannot delegate unvested coins. coins vested < delegation amount (%s < %s)",
				vested, delegateMsg.Amount.Amount,
			)
		}
	}

	return nil
}
