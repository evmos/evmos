// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/evmos/v19/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v19/x/feemarket/types"
)

// FeeMarketWrapper is a wrapper around the feemarket keeper
// that is used to manage an evm denom with 6 or 18 decimals.
// The wrapper makes the corresponding conversions to achieve:
//   - With the EVM, the wrapper works always with 18 decimals.
//   - With the feemarket module, the wrapper works always
//     with the bank module decimals (either 6 or 18).
type FeeMarketWrapper struct {
	types.FeeMarketKeeper
	decimals uint32
}

// NewFeeMarketWrapper creates a new feemarket Keeper wrapper instance.
// The BankWrapper is used to manage an evm denom with 6 or 18 decimals
func NewFeeMarketWrapper(
	fk types.FeeMarketKeeper,
) *FeeMarketWrapper {
	return &FeeMarketWrapper{
		fk,
		types.DefaultDenomDecimals,
	}
}

// WithDecimals function updates the decimals on the feemarket wrapper
// This function is useful when updating the evm params (denomDecimals)
func (w *FeeMarketWrapper) WithDecimals(decimals uint32) error {
	if decimals != types.Denom18Dec && decimals != types.Denom6Dec {
		// In historical queries, this will fail
		// So set 18 decimals as fallback
		w.decimals = types.Denom18Dec
		return fmt.Errorf("decimals = %d not supported. Valid values are %d and %d", decimals, types.Denom18Dec, types.Denom6Dec)
	}
	w.decimals = decimals
	return nil
}

// GetBaseFee returns the base fee converted to 18 decimals
func (w FeeMarketWrapper) GetBaseFee(ctx sdk.Context) *big.Int {
	baseFee := w.FeeMarketKeeper.GetBaseFee(ctx)
	if w.decimals == types.Denom18Dec {
		return baseFee.TruncateInt().BigInt()
	}
	return types.Convert6To18DecimalsLegacyDec(baseFee).TruncateInt().BigInt()
}

// CalculateBaseFee returns the calculated base fee converted to 18 decimals
func (w FeeMarketWrapper) CalculateBaseFee(ctx sdk.Context) *big.Int {
	baseFee := w.FeeMarketKeeper.CalculateBaseFee(ctx)
	if baseFee.IsNil() {
		return nil
	}
	if w.decimals == types.Denom18Dec {
		return baseFee.TruncateInt().BigInt()
	}
	return types.Convert6To18DecimalsLegacyDec(baseFee).TruncateInt().BigInt()
}

// GetParams returns the params converted to 18 decimals
func (w FeeMarketWrapper) GetParams(ctx sdk.Context) feemarkettypes.Params {
	params := w.FeeMarketKeeper.GetParams(ctx)
	if w.decimals == types.Denom6Dec {
		params.BaseFee = types.Convert6To18DecimalsLegacyDec(params.BaseFee)
		params.MinGasPrice = types.Convert6To18DecimalsLegacyDec(params.MinGasPrice)
	}
	return params
}
