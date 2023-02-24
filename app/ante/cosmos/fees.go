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

package cosmos

import (
	"fmt"
	"math"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/evmos/evmos/v11/app/ante/evm"
)

// DeductFeeDecorator deducts fees from the first signer of the tx.
// If the first signer does not have the funds to pay for the fees,
// and does not have enough unclaimed staking rewards, then return
// with InsufficientFunds error.
// The next AnteHandler is called if fees are successfully deducted.
//
// CONTRACT: Tx must implement FeeTx interface to use DeductFeeDecorator
type DeductFeeDecorator struct {
	accountKeeper      authante.AccountKeeper
	bankKeeper         BankKeeper
	distributionKeeper evm.DistributionKeeper
	feegrantKeeper     authante.FeegrantKeeper
	stakingKeeper      evm.StakingKeeper
	txFeeChecker       authante.TxFeeChecker
}

// NewDeductFeeDecorator returns a new DeductFeeDecorator.
func NewDeductFeeDecorator(
	ak authante.AccountKeeper,
	bk BankKeeper,
	dk evm.DistributionKeeper,
	fk authante.FeegrantKeeper,
	sk evm.StakingKeeper,
	tfc authante.TxFeeChecker,
) DeductFeeDecorator {
	if tfc == nil {
		tfc = checkTxFeeWithValidatorMinGasPrices
	}

	return DeductFeeDecorator{
		accountKeeper:      ak,
		bankKeeper:         bk,
		distributionKeeper: dk,
		feegrantKeeper:     fk,
		stakingKeeper:      sk,
		txFeeChecker:       tfc,
	}
}

// AnteHandle ensures that the transaction contains valid fee requirements and tries to deduct those
// from the account balance or unclaimed staking rewards, which the transaction sender might have.
func (dfd DeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(errortypes.ErrTxDecode, "Tx must be a FeeTx")
	}

	if !simulate && ctx.BlockHeight() > 0 && feeTx.GetGas() == 0 {
		return ctx, errorsmod.Wrap(errortypes.ErrInvalidGasLimit, "must provide positive gas")
	}

	var (
		priority int64
		err      error
	)

	fee := feeTx.GetFee()
	if !simulate {
		fee, priority, err = dfd.txFeeChecker(ctx, tx)
		if err != nil {
			return ctx, err
		}
	}
	if err := dfd.checkDeductFee(ctx, tx, fee); err != nil {
		return ctx, err
	}

	newCtx := ctx.WithPriority(priority)

	return next(newCtx, tx, simulate)
}

// checkDeductFee checks if the fee payer has enough funds to pay for the fees and deducts them.
// If the spendable balance is not enough, it tries to claim enough staking rewards to cover the fees.
func (dfd DeductFeeDecorator) checkDeductFee(ctx sdk.Context, sdkTx sdk.Tx, fees sdk.Coins) error {
	feeTx, ok := sdkTx.(sdk.FeeTx)
	if !ok {
		return errorsmod.Wrap(errortypes.ErrTxDecode, "Tx must implement the FeeTx interface")
	}

	if addr := dfd.accountKeeper.GetModuleAddress(authtypes.FeeCollectorName); addr == nil {
		return fmt.Errorf("fee collector module account (%s) has not been set", authtypes.FeeCollectorName)
	}

	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()
	deductFeesFrom := feePayer

	// if feegranter is set, then deduct the fee from the feegranter account.
	// this works only when feegrant is enabled.
	if feeGranter != nil {
		if dfd.feegrantKeeper == nil {
			return errortypes.ErrInvalidRequest.Wrap("fee grants are not enabled")
		} else if !feeGranter.Equals(feePayer) {
			err := dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranter, feePayer, fees, sdkTx.GetMsgs())
			if err != nil {
				return errorsmod.Wrapf(err, "%s does not not allow to pay fees for %s", feeGranter, feePayer)
			}
		}

		deductFeesFrom = feeGranter
	}

	deductFeesFromAcc := dfd.accountKeeper.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return errortypes.ErrUnknownAddress.Wrapf("fee payer address: %s does not exist", deductFeesFrom)
	}

	// deduct the fees
	if !fees.IsZero() {
		if err := dfd.deductFeesFromBalanceOrUnclaimedRewards(ctx, deductFeesFromAcc, fees); err != nil {
			return err
		}
	}

	events := sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeTx,
			sdk.NewAttribute(sdk.AttributeKeyFee, fees.String()),
			sdk.NewAttribute(sdk.AttributeKeyFeePayer, deductFeesFrom.String()),
		),
	}
	ctx.EventManager().EmitEvents(events)

	return nil
}

// deductFeesFromBalanceOrUnclaimedRewards tries to deduct the fees from the account balance.
// If the account balance is not enough, it tries to claim enough staking rewards to cover the fees.
// If the account does not have sufficient staking rewards, it returns an error.
func (dfd DeductFeeDecorator) deductFeesFromBalanceOrUnclaimedRewards(
	ctx sdk.Context,
	account authtypes.AccountI,
	fees sdk.Coins,
) error {
	stakingDenom := dfd.stakingKeeper.BondDenom(ctx)
	balance := dfd.bankKeeper.GetBalance(ctx, account.GetAddress(), stakingDenom)
	if !balance.IsPositive() {
		return errortypes.ErrInsufficientFunds.Wrapf("balance of %s in %s is not positive", account.GetAddress(), stakingDenom)
	}

	found, feesInStakingDenom := fees.Find(stakingDenom)
	if found && balance.IsLT(feesInStakingDenom) {
		difference := feesInStakingDenom.Sub(balance)
		// Try to claim enough staking rewards to cover the difference between the
		// transaction cost and the account balance.
		err := evm.ClaimSufficientStakingRewards(ctx, dfd.stakingKeeper, dfd.distributionKeeper, account.GetAddress(), sdk.Coins{difference})
		if err != nil {
			return errortypes.ErrInsufficientFunds.Wrapf(
				"insufficient funds and failed to claim sufficient staking rewards for %s to pay for fees; %s",
				account.GetAddress(),
				err.Error(),
			)
		}
	}

	// deduct the fees if possible
	return authante.DeductFees(dfd.bankKeeper, ctx, account, fees)
}

// checkTxFeeWithValidatorMinGasPrices implements the default fee logic, where the minimum price per
// unit of gas is fixed and set by each validator, and the tx priority is computed from the gas price.
func checkTxFeeWithValidatorMinGasPrices(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, errorsmod.Wrap(errortypes.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()

	// Ensure that the provided fees meets a minimum threshold for the validator,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on CheckTx.
	if ctx.IsCheckTx() {
		minGasPrices := ctx.MinGasPrices()
		if !minGasPrices.IsZero() {
			requiredFees := make(sdk.Coins, len(minGasPrices))

			// Determine the required fees by multiplying each required minimum gas
			// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
			glDec := sdk.NewDec(int64(gas)) //#nosec G701 -- gosec warning about integer overflow is not relevant here
			for i, gp := range minGasPrices {
				fee := gp.Amount.Mul(glDec)
				requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
			}

			if !feeCoins.IsAnyGTE(requiredFees) {
				return nil, 0, errorsmod.Wrapf(errortypes.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFees)
			}
		}
	}

	priority := getTxPriority(feeCoins, int64(gas)) //#nosec G701 -- gosec warning about integer overflow is not relevant here
	return feeCoins, priority, nil
}

// getTxPriority returns a naive tx priority based on the amount of the smallest denomination of the gas price
// provided in a transaction.
// NOTE: This implementation should be used with a great consideration as it opens potential attack vectors
// where txs with multiple coins could not be prioritized as expected.
func getTxPriority(fees sdk.Coins, gas int64) int64 {
	var priority int64
	for _, c := range fees {
		p := int64(math.MaxInt64)
		gasPrice := c.Amount.QuoRaw(gas)
		if gasPrice.IsInt64() {
			p = gasPrice.Int64()
		}
		if priority == 0 || p < priority {
			priority = p
		}
	}

	return priority
}
