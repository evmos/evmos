// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v20/x/evm/types"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// BeginBlock emits a base fee event which will be adjusted to the evm decimals
func (k *Keeper) BeginBlock(ctx sdk.Context) error {
	logger := ctx.Logger().With("begin_block", "evm")

	// Base fee is already set on FeeMarket BeginBlock
	// that runs before this one
	// We emit this event on the EVM and FeeMarket modules
	// because they can be different if the evm denom has 6 decimals
	res, err := k.BaseFee(ctx, &types.QueryBaseFeeRequest{})
	if err != nil {
		logger.Error("error when getting base fee", "error", err.Error())
	}
	if res != nil && res.BaseFee != nil && !res.BaseFee.IsNil() {
		// Store current base fee in event
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeFeeMarket,
				sdk.NewAttribute(types.AttributeKeyBaseFee, res.BaseFee.String()),
			),
		})
	}
	return nil
}

// EndBlock also retrieves the bloom filter value from the transient store and commits it to the
// KVStore. The EVM end block logic doesn't update the validator set, thus it returns
// an empty slice.
func (k *Keeper) EndBlock(ctx sdk.Context) error {
	// Gas costs are handled within msg handler so costs should be ignored
	infCtx := ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())

	bloom := ethtypes.BytesToBloom(k.GetBlockBloomTransient(infCtx).Bytes())
	k.EmitBlockBloomEvent(infCtx, bloom)

	return nil
}
