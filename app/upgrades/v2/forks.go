package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	erc20keeper "github.com/tharsis/evmos/x/erc20/keeper"
)

func RunForkLogic(ctx sdk.Context, erc20 *erc20keeper.Keeper) {
	ctx.Logger().Info("Applying Evmos v2 upgrade. Setting ERC20 module evmhook to true")

	FixErc20Param(ctx, erc20)
}

func FixErc20Param(ctx sdk.Context, erc20 *erc20keeper.Keeper) {
	// update erc20 evmhook param to true
	params := erc20.GetParams(ctx)
	params.EnableEVMHook = true
	erc20.SetParams(ctx, params)
}
