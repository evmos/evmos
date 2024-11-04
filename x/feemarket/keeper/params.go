// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package keeper

import (
	"cosmossdk.io/math"
	"github.com/evmos/evmos/v20/x/feemarket/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams returns the total set of fee market parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if len(bz) == 0 {
		k.ss.GetParamSetIfExists(ctx, &params)
	} else {
		k.cdc.MustUnmarshal(bz, &params)
	}

	// zero the nil params for legacy blocks
	if params.MinGasPrice.IsNil() {
		params.MinGasPrice = math.LegacyZeroDec()
	}

	if params.MinGasMultiplier.IsNil() {
		params.MinGasMultiplier = math.LegacyZeroDec()
	}

	return
}

// SetParams sets the fee market params in a single key
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(types.ParamsKey, bz)

	return nil
}

// ----------------------------------------------------------------------------
// Parent Base Fee
// Required by EIP1559 base fee calculation.
// ----------------------------------------------------------------------------

// GetBaseFeeEnabled returns true if base fee is enabled
func (k Keeper) GetBaseFeeEnabled(ctx sdk.Context) bool {
	params := k.GetParams(ctx)
	return !params.NoBaseFee && ctx.BlockHeight() >= params.EnableHeight
}

// GetBaseFee gets the base fee from the store
func (k Keeper) GetBaseFee(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	if params.NoBaseFee {
		return math.LegacyDec{}
	}

	baseFee := params.BaseFee
	if baseFee.IsNil() || baseFee.IsZero() {
		bfV1 := k.GetBaseFeeV1(ctx)
		if bfV1 == nil {
			return math.LegacyDec{}
		}
		// try v1 format
		return math.LegacyNewDecFromBigInt(bfV1)
	}
	return baseFee
}

// SetBaseFee set's the base fee in the store
func (k Keeper) SetBaseFee(ctx sdk.Context, baseFee math.LegacyDec) {
	params := k.GetParams(ctx)
	params.BaseFee = baseFee
	err := k.SetParams(ctx, params)
	if err != nil {
		return
	}
}
