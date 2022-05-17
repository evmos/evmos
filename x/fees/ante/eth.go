package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
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

func (empd EthMinPriceFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	evmDenom := empd.evmKeeper.GetParams(ctx).EvmDenom
	minGasPrice := empd.feesKeeper.GetParams(ctx).MinGasPrice
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

			// For dynamic transactions, GetFee() uses the GasFeeCap value, which
			// is the maximum gas price that the signer can pay. In practice, the
			// signer can pay less, if the block's BaseFee is lower. So, in this case,
			// we use the EffectiveFee. If the feemarket formula results in a BaseFee
			// that lowers EffectivePrice until it is < MinGasPrices, the users must
			// increase the GasTipCap (priority fee) until EffectivePrice > MinGasPrices.
			// Transactions with MinGasPrices * gasUsed < tx fees < EffectiveFee are rejected
			// by the feemarket AnteHandle
			txData, err := evmtypes.UnpackTxData(ethMsg.Data)
			if err == nil && txData.TxType() != ethtypes.LegacyTxType {
				paramsEvm := empd.evmKeeper.GetParams(ctx)
				ethCfg := paramsEvm.ChainConfig.EthereumConfig(empd.evmKeeper.ChainID())
				baseFee := empd.evmKeeper.GetBaseFee(ctx, ethCfg)
				feeAmt = ethMsg.GetEffectiveFee(baseFee)
			}

			glDec := sdk.NewDec(int64(ethMsg.GetGas()))
			requiredFee := minGasPrices.AmountOf(evmDenom).Mul(glDec)

			if sdk.NewDecFromBigInt(feeAmt).LT(requiredFee) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "gas price less than fees module MinGasPrices; got: %s required: %s", feeAmt, requiredFee)
			}
		}
	}

	return next(ctx, tx, simulate)
}
