// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package cosmos

import (
	"fmt"
	"math/big"

	"golang.org/x/exp/slices"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	evmante "github.com/evmos/evmos/v16/app/ante/evm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
)

// MinGasPriceDecorator will check if the transaction's fee is at least as large
// as the MinGasPrices param. If fee is too low, decorator returns error and tx
// is rejected. This applies for both CheckTx and DeliverTx
// If fee is high enough, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use MinGasPriceDecorator
type MinGasPriceDecorator struct {
	feesKeeper evmante.FeeMarketKeeper
	evmKeeper  evmante.EVMKeeper
}

// NewMinGasPriceDecorator creates a new MinGasPriceDecorator instance used only for
// Cosmos transactions.
func NewMinGasPriceDecorator(fk evmante.FeeMarketKeeper, ek evmante.EVMKeeper) MinGasPriceDecorator {
	return MinGasPriceDecorator{feesKeeper: fk, evmKeeper: ek}
}

func (mpd MinGasPriceDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrapf(errortypes.ErrInvalidType, "invalid transaction type %T, expected sdk.FeeTx", tx)
	}

	minGasPrice := mpd.feesKeeper.GetParams(ctx).MinGasPrice

	feeCoins := feeTx.GetFee()
	evmParams := mpd.evmKeeper.GetParams(ctx)
	evmDenom := evmParams.GetEvmDenom()

	// only allow user to pass in aevmos and stake native token as transaction fees
	// allow use stake native tokens for fees is just for unit tests to pass
	validFees := len(feeCoins) == 0 || (len(feeCoins) == 1 && slices.Contains([]string{evmDenom, sdk.DefaultBondDenom}, feeCoins.GetDenomByIndex(0)))
	if !validFees && !simulate {
		return ctx, fmt.Errorf("expected only use native token %s for fee, but got %s", evmDenom, feeCoins.String())
	}

	// Short-circuit if min gas price is 0 or if simulating
	if minGasPrice.IsZero() || simulate {
		return next(ctx, tx, simulate)
	}

	minGasPrices := sdk.DecCoins{
		{
			Denom:  evmDenom,
			Amount: minGasPrice,
		},
	}

	gas := feeTx.GetGas()

	requiredFees := make(sdk.Coins, 0)

	// Determine the required fees by multiplying each required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	gasLimit := math.LegacyNewDecFromBigInt(new(big.Int).SetUint64(gas))

	for _, gp := range minGasPrices {
		fee := gp.Amount.Mul(gasLimit).Ceil().RoundInt()
		if fee.IsPositive() {
			requiredFees = requiredFees.Add(sdk.Coin{Denom: gp.Denom, Amount: fee})
		}
	}

	// Fees not provided (or flag "auto"). Then use the base fee to make the check pass
	if feeCoins == nil {
		return ctx, errorsmod.Wrapf(errortypes.ErrInsufficientFee,
			"fee not provided. Please use the --fees flag or the --gas-price flag along with the --gas flag to estimate the fee. The minimum global fee for this tx is: %s",
			requiredFees)
	}

	if !feeCoins.IsAnyGTE(requiredFees) {
		return ctx, errorsmod.Wrapf(errortypes.ErrInsufficientFee,
			"provided fee < minimum global fee (%s < %s). Please increase the gas price.",
			feeCoins,
			requiredFees)
	}

	return next(ctx, tx, simulate)
}
