package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v2 "github.com/tharsis/evmos/app/upgrades/v2"
)

// BeginBlockForks executes any necessary fork logic based upon the current block height.
func BeginBlockForks(ctx sdk.Context, app *Evmos) {
	switch ctx.BlockHeight() {
	case v2.UpgradeHeight:
		v2.RunForkLogic(ctx, &app.Erc20Keeper)
	default:
		// do nothing
		return
	}
}
