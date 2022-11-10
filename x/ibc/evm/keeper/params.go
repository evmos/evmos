package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v9/x/ibc/evm/types"

)

// GetSendEnabled retrieves the send enabled boolean from the paramstore
func (k Keeper) GetSendEvmTxEnabled(ctx sdk.Context) bool {
	var res bool
	k.paramSpace.Get(ctx, types.KeySendEvmTxEnabled, &res)
	return res
}

// GetReceiveEnabled retrieves the receive enabled boolean from the paramstore
func (k Keeper) GetReceiveEvmTxEnabled(ctx sdk.Context) bool {
	var res bool
	k.paramSpace.Get(ctx, types.KeyReceiveEvmTxEnabled, &res)
	return res
}

// GetParams returns the total set of ibc-transfer parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(k.GetSendEvmTxEnabled(ctx), k.GetReceiveEvmTxEnabled(ctx))
}

// SetParams sets the total set of ibc-transfer parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
