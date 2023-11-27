// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/evmos/evmos/v15/types"
)

// CheckGasWanted keeps track of the gasWanted amount on the current block in transient store
// for BaseFee calculation.
func CheckGasWanted(ctx sdk.Context, feeMarketKeeper FeeMarketKeeper, tx sdk.Tx, isLondon bool) error {
	if !isLondon {
		return nil
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil
	}

	gasWanted := feeTx.GetGas()

	// return error if the tx gas is greater than the block limit (max gas)
	blockGasLimit := types.BlockGasLimit(ctx)
	if gasWanted > blockGasLimit {
		return errorsmod.Wrapf(
			errortypes.ErrOutOfGas,
			"tx gas (%d) exceeds block gas limit (%d)",
			gasWanted,
			blockGasLimit,
		)
	}

	isBaseFeeEnabled := feeMarketKeeper.GetBaseFeeEnabled(ctx)
	if !isBaseFeeEnabled {
		return nil
	}

	// Add total gasWanted to cumulative in block transientStore in FeeMarket module
	if _, err := feeMarketKeeper.AddTransientGasWanted(ctx, gasWanted); err != nil {
		return errorsmod.Wrapf(err, "failed to add gas wanted to transient store")
	}

	return nil
}
