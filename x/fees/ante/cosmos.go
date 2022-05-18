package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MinPriceFeeDecorator will check if the transaction's fee is at least as large
// as the MinGasPrices param. If fee is too low, decorator returns error and tx
// is rejected. This applies for both CheckTx and DeliverTx
// If fee is high enough, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use MinPriceFeeDecorator
type MinPriceFeeDecorator struct {
	feesKeeper FeesKeeper
	evmKeeper  EvmKeeper
}

func NewMinPriceFeeDecorator(fk FeesKeeper, ek EvmKeeper) MinPriceFeeDecorator {
	return MinPriceFeeDecorator{feesKeeper: fk, evmKeeper: ek}
}

func (mpd MinPriceFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	minGasPrice := mpd.feesKeeper.GetParams(ctx).MinGasPrice
	minGasPrices := sdk.DecCoins{sdk.DecCoin{
		Denom:  mpd.evmKeeper.GetParams(ctx).EvmDenom,
		Amount: minGasPrice,
	}}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()

	if !minGasPrices.IsZero() {
		requiredFees := make(sdk.Coins, len(minGasPrices))

		// Determine the required fees by multiplying each required minimum gas
		// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
		glDec := sdk.NewDec(int64(gas))
		for i, gp := range minGasPrices {
			fee := gp.Amount.Mul(glDec)
			requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}

		if !feeCoins.IsAnyGTE(requiredFees) {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "provided fee < minimum global fee (%s < %s). Please increase the gas price.", feeCoins, requiredFees)
		}
	}

	return next(ctx, tx, simulate)
}
