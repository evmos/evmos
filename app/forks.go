package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v2 "github.com/tharsis/evmos/app/upgrades/v2"
)

// BeginBlockForks is intended to be ran in
func BeginBlockForks(ctx sdk.Context, app *Evmos) {
	switch ctx.BlockHeight() {
	case v2.UpgradeHeight:
		v2.RunForkLogic(ctx, &app.Erc20Keeper)
	default:
		// do nothing
		return
	}
}
