package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/evmos/v3/x/recovery/types"
)

// GetParams returns the total set of incentives parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the incentives parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
