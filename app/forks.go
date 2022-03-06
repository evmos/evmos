package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	v2 "github.com/tharsis/evmos/app/upgrades/v2"
)

// BeginBlockForks executes any necessary fork logic based upon the current block height.
func BeginBlockForks(ctx sdk.Context, app *Evmos) {
	switch ctx.BlockHeight() {
	case v2.UpgradeHeight:
		upgradePlan := upgradetypes.Plan{
			Name:   v2.UpgradeName,
			Info:   v2.UpgradeInfo,
			Height: v2.UpgradeHeight,
		}
		err := app.UpgradeKeeper.ScheduleUpgradeNoHeightCheck(ctx, upgradePlan)
		if err != nil {
			panic(err)
		}
	default:
		// do nothing
		return
	}
}
