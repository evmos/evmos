package app

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	v2 "github.com/tharsis/evmos/v3/app/upgrades/mainnet/v2"
	v4 "github.com/tharsis/evmos/v3/app/upgrades/mainnet/v4"
	tv4 "github.com/tharsis/evmos/v3/app/upgrades/testnet/v4"
)

// BeginBlockForks executes any necessary fork logic for based upon the current
// block height.
func BeginBlockForks(ctx sdk.Context, app *Evmos) {
	if strings.HasPrefix(ctx.ChainID(), MainnetChainID) {
		scheduleMainnetUpgrades(ctx, app)
	} else if strings.HasPrefix(ctx.ChainID(), TestnetChainID) {
		scheduleTestnetUpgrades(ctx, app)
	}
}

func scheduleMainnetUpgrades(ctx sdk.Context, app *Evmos) {
	upgradePlan := upgradetypes.Plan{}

	switch ctx.BlockHeight() {
	case v2.UpgradeHeight:
		upgradePlan = upgradetypes.Plan{
			Name:   v2.UpgradeName,
			Info:   v2.UpgradeInfo,
			Height: v2.UpgradeHeight,
		}
	case v4.UpgradeHeight:
		upgradePlan = upgradetypes.Plan{
			Name:   v4.UpgradeName,
			Info:   v4.UpgradeInfo,
			Height: v4.UpgradeHeight,
		}
	default:
		return
	}

	err := app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan)
	if err != nil {
		panic(
			fmt.Errorf(
				"failed to schedule upgrade %s during BeginBlock at height %d: %w",
				upgradePlan.Name, ctx.BlockHeight(), err,
			),
		)
	}
}

func scheduleTestnetUpgrades(ctx sdk.Context, app *Evmos) {
	upgradePlan := upgradetypes.Plan{}

	switch ctx.BlockHeight() {
	case tv4.UpgradeHeight:
		upgradePlan = upgradetypes.Plan{
			Name:   tv4.UpgradeName,
			Info:   tv4.UpgradeInfo,
			Height: tv4.UpgradeHeight,
		}
	default:
		return
	}

	err := app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan)
	if err != nil {
		panic(
			fmt.Errorf(
				"failed to schedule upgrade %s during BeginBlock at height %d: %w",
				upgradePlan.Name, ctx.BlockHeight(), err,
			),
		)
	}
}
