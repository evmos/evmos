package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evoblockchain/evoblock/v8/x/inflation/types"
)

// GetParams returns the total set of inflation parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramstore.GetParamSetIfExists(ctx, &params)
	return params
}

// SetParams sets the inflation parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
