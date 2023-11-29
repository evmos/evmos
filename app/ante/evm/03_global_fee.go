// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
)

// EthMinGasPriceDecorator will check if the transaction's fee is at least as large
// as the MinGasPrices param. If fee is too low, decorator returns error and tx
// is rejected. This applies to both CheckTx and DeliverTx and regardless
// if London hard fork or fee market params (EIP-1559) are enabled.
// If fee is high enough, then call next AnteHandler
type EthMinGasPriceDecorator struct {
	feesKeeper FeeMarketKeeper
	evmKeeper  EVMKeeper
}

// NewEthMinGasPriceDecorator creates a new MinGasPriceDecorator instance used only for
// Ethereum transactions.
func NewEthMinGasPriceDecorator(fk FeeMarketKeeper, ek EVMKeeper) EthMinGasPriceDecorator {
	return EthMinGasPriceDecorator{feesKeeper: fk, evmKeeper: ek}
}

// AnteHandle ensures that the effective fee from the transaction is greater than the
// minimum global fee, which is defined by the  MinGasPrice (parameter) * GasLimit (tx argument).
func (empd EthMinGasPriceDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	minGasPrice := empd.feesKeeper.GetParams(ctx).MinGasPrice

	// short-circuit if min gas price is 0
	if minGasPrice.IsZero() {
		return next(ctx, tx, simulate)
	}

	evmParams := empd.evmKeeper.GetParams(ctx)
	chainCfg := evmParams.GetChainConfig()
	ethCfg := chainCfg.EthereumConfig(empd.evmKeeper.ChainID())
	baseFee := empd.evmKeeper.GetBaseFee(ctx, ethCfg)

	for _, msg := range tx.GetMsgs() {
		_, txData, _, err := evmtypes.UnpackEthMsg(msg)
		if err != nil {
			return ctx, err
		}

		feeAmt := txData.Fee()

		if txData.TxType() != ethtypes.LegacyTxType {
			feeAmt = txData.EffectiveFee(baseFee)
		}

		gasLimit := sdk.NewDecFromBigInt(new(big.Int).SetUint64(txData.GetGas()))
		fee := sdkmath.LegacyNewDecFromBigInt(feeAmt)

		if err := CheckGlobalFee(fee, minGasPrice, gasLimit); err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

// For dynamic transactions, GetFee() uses the GasFeeCap value, which
// is the maximum gas price that the signer can pay. In practice, the
// signer can pay less, if the block's BaseFee is lower. So, in this case,
// we use the EffectiveFee. If the feemarket formula results in a BaseFee
// that lowers EffectivePrice until it is < MinGasPrices, the users must
// increase the GasTipCap (priority fee) until EffectivePrice > MinGasPrices.
// Transactions with MinGasPrices * gasUsed < tx fees < EffectiveFee are rejected
// by the feemarket AnteHandle
func CheckGlobalFee(fee, globalMinGasPrice, gasLimit sdkmath.LegacyDec) error {
	if globalMinGasPrice.IsZero() {
		return nil
	}

	requiredFee := globalMinGasPrice.Mul(gasLimit)

	if fee.LT(requiredFee) {
		return errorsmod.Wrapf(
			errortypes.ErrInsufficientFee,
			"provided fee < minimum global fee (%s < %s). Please increase the priority tip (for EIP-1559 txs) or the gas prices (for access list or legacy txs)", //nolint:lll
			fee.TruncateInt().String(), requiredFee.TruncateInt().String(),
		)
	}

	return nil
}
