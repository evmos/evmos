package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v7/x/erc20/types"
)

// GetParams returns the total set of erc20 parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the erc20 parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
