package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
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
		app.UpgradeKeeper.ScheduleUpgradeNoHeightCheck(ctx, upgradePlan)
	default:
		// do nothing
		return
	}
}
