package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/evmos/x/erc20/types"
)

type ERC20Keeper interface {
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params)
}

func UpdateParams(ctx sdk.Context, k ERC20Keeper) error {
	params := k.GetParams(ctx)
	params.EnableErc20 = true
	params.EnableEVMHook = true
	k.SetParams(ctx, params)
	return nil
}
