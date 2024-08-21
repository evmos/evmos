package keeper

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

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

// WithDecimals function updates the decimals on the bank wrapper
// This function is useful when updating the evm params (denomDecimals)
func (w *FeeMarketWrapper) WithDecimals(decimals uint32) error {
	if decimals != types.Denom18Dec && decimals != types.Denom6Dec {
		return fmt.Errorf("decimals = %d not supported. Valid values are %d and %d", decimals, types.Denom18Dec, types.Denom6Dec)
	}
	w.decimals = decimals
	return nil
}

// GetBaseFee returns the base fee converted to 18 decimals
func (w FeeMarketWrapper) GetBaseFee(ctx sdk.Context) *big.Int {
	baseFee := w.FeeMarketKeeper.GetBaseFee(ctx)
	if w.decimals == types.Denom18Dec {
		return baseFee
	}
	return types.Convert6To18DecimalsBigInt(baseFee)
}

// CalculateBaseFee returns the calculated base fee converted to 18 decimals
func (w FeeMarketWrapper) CalculateBaseFee(ctx sdk.Context) *big.Int {
	baseFee := w.FeeMarketKeeper.CalculateBaseFee(ctx)
	if w.decimals == types.Denom18Dec {
		return baseFee
	}
	return types.Convert6To18DecimalsBigInt(baseFee)
}

// GetBaseFee returns the base fee converted to 18 decimals
func (w FeeMarketWrapper) GetParams(ctx sdk.Context) feemarkettypes.Params {
	params := w.FeeMarketKeeper.GetParams(ctx)
	if w.decimals == types.Denom6Dec {
		convertedBaseFee := types.Convert6To18DecimalsBigInt(params.BaseFee.BigInt())
		params.BaseFee = sdk.NewIntFromBigInt(convertedBaseFee)
		convertedMinGasPrice := types.Convert6To18DecimalsBigInt(params.MinGasPrice.BigInt())
		params.MinGasPrice = math.LegacyNewDecFromBigInt(convertedMinGasPrice)
	}
	return params
}
