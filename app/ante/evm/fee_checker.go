// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm

import (
	"math"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	evmostypes "github.com/evmos/evmos/v20/types"
	"github.com/evmos/evmos/v20/x/evm/types"
)

// NewDynamicFeeChecker returns a `TxFeeChecker` that applies a dynamic fee to
// Cosmos txs using the EIP-1559 fee market logic.
// This can be called in both CheckTx and deliverTx modes.
// a) feeCap = tx.fees / tx.gas
// b) tipFeeCap = tx.MaxPriorityPrice (default) or MaxInt64
// - when `ExtensionOptionDynamicFeeTx` is omitted, `tipFeeCap` defaults to `MaxInt64`.
// - when london hardfork is not enabled, it falls back to SDK default behavior (validator min-gas-prices).
// - Tx priority is set to `effectiveGasPrice / DefaultPriorityReduction`.
func NewDynamicFeeChecker(fmk FeeMarketKeeper) authante.TxFeeChecker {
	return func(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
		feeTx, ok := tx.(sdk.FeeTx)
		if !ok {
			return sdk.Coins{}, 0, errorsmod.Wrap(errortypes.ErrTxDecode, "Tx must be a FeeTx")
		}
		// TODO: in the e2e test, if the fee in the genesis transaction meet the baseFee and minGasPrice in the feemarket, we can remove this code
		if ctx.BlockHeight() == 0 {
			// genesis transactions: fallback to min-gas-price logic
			return checkTxFeeWithValidatorMinGasPrices(ctx, feeTx)
		}

		return feeChecker(ctx, fmk, feeTx)
	}
}

// feeChecker returns the effective fee and priority for a given transaction.
func feeChecker(
	ctx sdk.Context,
	k FeeMarketKeeper,
	feeTx sdk.FeeTx,
) (sdk.Coins, int64, error) {
	denom := types.GetEVMCoinDenom()
	ethConfig := types.GetChainConfig()

	if !types.IsLondon(ethConfig, ctx.BlockHeight()) {
		// london hardfork is not enabled: fallback to min-gas-prices logic
		return checkTxFeeWithValidatorMinGasPrices(ctx, feeTx)
	}

	baseFee := k.GetBaseFee(ctx)
	// if baseFee is nil because it is disabled
	// or not found, consider it as 0
	// so the DynamicFeeTx logic can be applied
	if baseFee.IsNil() {
		baseFee = sdkmath.LegacyZeroDec()
	}

	// default to `MaxInt64` when there's no extension option.
	maxPriorityPrice := sdkmath.LegacyNewDec(math.MaxInt64)

	// get the priority tip cap from the extension option.
	if hasExtOptsTx, ok := feeTx.(authante.HasExtensionOptionsTx); ok {
		for _, opt := range hasExtOptsTx.GetExtensionOptions() {
			if extOpt, ok := opt.GetCachedValue().(*evmostypes.ExtensionOptionDynamicFeeTx); ok {
				maxPriorityPrice = extOpt.MaxPriorityPrice
				if maxPriorityPrice.IsNil() {
					maxPriorityPrice = sdkmath.LegacyZeroDec()
				}
				break
			}
		}
	}

	// priority fee cannot be negative
	if maxPriorityPrice.IsNegative() {
		return nil, 0, errorsmod.Wrapf(errortypes.ErrInsufficientFee, "max priority price cannot be negative")
	}

	gas := sdkmath.NewIntFromUint64(feeTx.GetGas())
	if gas.IsZero() {
		return nil, 0, errorsmod.Wrap(errortypes.ErrInvalidRequest, "gas cannot be zero")
	}

	feeCoins := feeTx.GetFee()
	feeAmtDec := sdkmath.LegacyNewDecFromInt(feeCoins.AmountOfNoDenomValidation(denom))

	feeCap := feeAmtDec.QuoInt(gas)

	if feeCap.LT(baseFee) {
		return nil, 0, errorsmod.Wrapf(errortypes.ErrInsufficientFee, "gas prices too low, got: %s%s required: %s%s. Please retry using a higher gas price or a higher fee", feeCap, denom, baseFee, denom)
	}

	// calculate the effective gas price using the EIP-1559 logic.
	effectivePrice := effectiveGasPriceLegacyDec(baseFee, feeCap, maxPriorityPrice)

	// NOTE: create a new coins slice without having to validate the denom
	effectiveFee := sdk.Coins{
		{
			Denom:  denom,
			Amount: effectivePrice.MulInt(gas).Ceil().RoundInt(),
		},
	}

	priorityInt := effectivePrice.Sub(baseFee).QuoInt(types.DefaultPriorityReduction).TruncateInt()
	priority := int64(math.MaxInt64)

	if priorityInt.IsInt64() {
		priority = priorityInt.Int64()
	}

	return effectiveFee, priority, nil
}

// checkTxFeeWithValidatorMinGasPrices implements the default fee logic, where the minimum price per
// unit of gas is fixed and set by each validator, and the tx priority is computed from the gas price.
func checkTxFeeWithValidatorMinGasPrices(ctx sdk.Context, tx sdk.FeeTx) (sdk.Coins, int64, error) {
	feeCoins := tx.GetFee()
	minGasPrices := ctx.MinGasPrices()
	gas := int64(tx.GetGas()) //#nosec G701 G115 -- checked for int overflow on ValidateBasic()

	// Ensure that the provided fees meet a minimum threshold for the validator,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	if ctx.IsCheckTx() && !minGasPrices.IsZero() {
		requiredFees := make(sdk.Coins, len(minGasPrices))

		// Determine the required fees by multiplying each required minimum gas
		// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
		glDec := sdkmath.LegacyNewDec(gas)
		for i, gp := range minGasPrices {
			fee := gp.Amount.Mul(glDec)
			requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}

		if !feeCoins.IsAnyGTE(requiredFees) {
			return nil, 0, errorsmod.Wrapf(errortypes.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFees)
		}
	}

	priority := getTxPriority(feeCoins, gas)
	return feeCoins, priority, nil
}

// getTxPriority returns a naive tx priority based on the amount of the smallest denomination of the gas price
// provided in a transaction.
func getTxPriority(fees sdk.Coins, gas int64) int64 {
	var priority int64

	for _, fee := range fees {
		gasPrice := fee.Amount.QuoRaw(gas)
		amt := gasPrice.Quo(types.DefaultPriorityReduction)
		p := int64(math.MaxInt64)

		if amt.IsInt64() {
			p = amt.Int64()
		}

		if priority == 0 || p < priority {
			priority = p
		}
	}

	return priority
}

// effectiveGasPriceLegacyDec computes the effective gas price based on eip-1559 rules
// `effectiveGasPrice = min(baseFee + tipCap, feeCap)` using decimals
func effectiveGasPriceLegacyDec(baseFee, feeCap, tipCap sdkmath.LegacyDec) sdkmath.LegacyDec {
	return sdkmath.LegacyMinDec(tipCap.Add(baseFee), feeCap)
}
