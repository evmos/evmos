package app

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	v2 "github.com/tharsis/evmos/v4/app/upgrades/mainnet/v2"
	tv3 "github.com/tharsis/evmos/v4/app/upgrades/testnet/v3"
)

// BeginBlockForks executes any necessary fork logic based upon the current block height.
func BeginBlockForks(ctx sdk.Context, app *Evmos) {
	switch ctx.BlockHeight() {
	case v2.UpgradeHeight:
		// NOTE: only run for mainnet
		if !strings.HasPrefix(ctx.ChainID(), MainnetChainID) {
			return
		}

		upgradePlan := upgradetypes.Plan{
			Name:   v2.UpgradeName,
			Info:   v2.UpgradeInfo,
			Height: v2.UpgradeHeight,
		}

		err := app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan)
		if err != nil {
			panic(err)
		}

	// NOTE: THIS UPGRADE PLAN SHOULD BE ADDED TO VERSION 1.0.0-beta1
	// This will create the upgrade plan on the defined height
	// and will stop the chain once the height is reached.
	// The new version should have the upgrade migration logic under the upgradeName
	case tv3.UpgradeHeight:
		// NOTE: only run for testnet
		if !strings.HasPrefix(ctx.ChainID(), TestnetChainID) {
			return
		}

		upgradePlan := upgradetypes.Plan{
			Name:   tv3.UpgradeName,
			Info:   tv3.UpgradeInfo,
			Height: tv3.UpgradeHeight + 1,
		}
		err := app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan)
		if err != nil {
			panic(err)
		}
	default:
		// do nothing
		return
	}
}
