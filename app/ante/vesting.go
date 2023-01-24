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
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	evmostypes "github.com/evmos/evmos/v11/types"
	vestingtypes "github.com/evmos/evmos/v11/x/vesting/types"
)

// EthVestingTransactionDecorator validates if clawback vesting accounts are
// permitted to perform Ethereum Tx.
type EthVestingTransactionDecorator struct {
	ak evmtypes.AccountKeeper
	bk evmtypes.BankKeeper
}

func NewEthVestingTransactionDecorator(ak evmtypes.AccountKeeper, bk evmtypes.BankKeeper) EthVestingTransactionDecorator {
	return EthVestingTransactionDecorator{
		ak: ak,
		bk: bk,
	}
}

// AnteHandle validates that a clawback vesting account has surpassed the
// vesting cliff and lockup period.
//
// This AnteHandler decorator will fail if:
//   - the message is not a MsgEthereumTx
//   - sender account cannot be found
//   - tx values are in excess of any account's spendable balances
//   - blocktime is before surpassing vesting cliff end (with zero vested coins)
func (vtd EthVestingTransactionDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// Track the total `value` to be spent by each address across all messages and ensure
	// that no cumulative `value` exceeds any account's spendable balance.
	totalValueByAddress := make(map[string]sdk.Coin)

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
			return next(ctx, tx, simulate)
		}

		// Error if vesting cliff has not passed (with zero vested coins). This
		// rule does not apply for existing clawback accounts that receive a new
		// grant while there are already vested coins on the account.
		vested := clawbackAccount.GetVestedCoins(ctx.BlockTime())
		if len(vested) == 0 {
			return ctx, errorsmod.Wrapf(vestingtypes.ErrInsufficientVestedCoins,
				"cannot perform Ethereum tx with clawback vesting account, that has no vested coins: %s", vested,
			)
		}

		msgValue := sdk.NewCoin(evmostypes.BaseDenom, math.NewIntFromBigInt(msgEthTx.AsTransaction().Value()))
		address := acc.GetAddress()

		// Since there can be multiple transactions from different accounts, we track each account's total
		// requested value to compare against its unlocked balances.
		totalValue, ok := totalValueByAddress[address.String()]
		if !ok {
			totalValue = msgValue
		} else {
			totalValue = totalValue.Add(msgValue)
		}
		totalValueByAddress[address.String()] = totalValue

		// Check that the clawbackAccount has suffient unlocked tokens to cover all requested spending
		// lockedBalance defaults to zero if not found.
		_, lockedBalance := clawbackAccount.LockedCoins(ctx.BlockTime()).Find(evmostypes.BaseDenom)
		spendableBalance, err := vtd.bk.GetBalance(ctx, address, evmostypes.BaseDenom).SafeSub(lockedBalance)
		if err != nil {
			spendableBalance = sdk.NewCoin(evmostypes.BaseDenom, sdk.ZeroInt())
		}

		if totalValue.Amount.GT(spendableBalance.Amount) {
			return ctx, errorsmod.Wrapf(vestingtypes.ErrInsufficientUnlockedCoins,
				"clawback vesting account has insufficient unlocked tokens to execute transaction: %s", spendableBalance,
			)
		}
	}

	return next(ctx, tx, simulate)
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
