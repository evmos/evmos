package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/evmos/v3/x/claims/types"
)

// GetParams returns the total set of claim parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the claim parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
