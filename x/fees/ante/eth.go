package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

// EthMinPriceFeeDecorator will check if the transaction's fee is at least as large
// as the MinGasPrices param. If fee is too low, decorator returns error and tx
// is rejected. This applies for both CheckTx and DeliverTx. This applies regardless
// if london fork or feemarket are enabled
// If fee is high enough, then call next AnteHandler
type EthMinPriceFeeDecorator struct {
	feesKeeper FeesKeeper
	evmKeeper  EvmKeeper
}

func NewEthMinPriceFeeDecorator(fk FeesKeeper, ek EvmKeeper) EthMinPriceFeeDecorator {
	return EthMinPriceFeeDecorator{feesKeeper: fk, evmKeeper: ek}
}

func (mpd EthMinPriceFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	evmDenom := mpd.evmKeeper.GetParams(ctx).EvmDenom
	minGasPrice := mpd.feesKeeper.GetParams(ctx).MinGasPrice
	minGasPrices := sdk.DecCoins{sdk.DecCoin{
		Denom:  evmDenom,
		Amount: minGasPrice,
	}}

	if !minGasPrices.IsZero() {
		for _, msg := range tx.GetMsgs() {
			ethMsg, ok := msg.(*evmtypes.MsgEthereumTx)
			if !ok {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid message type %T, expected %T", msg, (*evmtypes.MsgEthereumTx)(nil))
			}

			feeAmt := ethMsg.GetFee()
			glDec := sdk.NewDec(int64(ethMsg.GetGas()))
			requiredFee := minGasPrices.AmountOf(evmDenom).Mul(glDec)

			if sdk.NewDecFromBigInt(feeAmt).LT(requiredFee) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "gas price less than fees module MinGasPrices; got: %s required: %s", feeAmt, requiredFee)
			}
		}
	}

	return next(ctx, tx, simulate)
}
