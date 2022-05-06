package app

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	v2 "github.com/tharsis/evmos/v4/app/upgrades/v2"
	v4 "github.com/tharsis/evmos/v4/app/upgrades/v4"
)

// BeginBlockForks executes any necessary fork logic for based upon the current
// block height and chain ID (mainnet or testnet). It sets an upgrade plan once
// the chain reaches the pre-defined upgrade height.
//
// CONTRACT: for this logic to work properly it is required to:
//
// 	1) Release a non-breaking patch version so that the chain can set the scheduled upgrade plan at upgrade-height.
// 	2) Release the software defined in the upgrade-info
func BeginBlockForks(ctx sdk.Context, app *Evmos) {
	upgradePlan := upgradetypes.Plan{
		Height: ctx.BlockHeight(),
	}

	if strings.HasPrefix(ctx.ChainID(), MainnetChainID) {
		// handle mainnet forks
		switch ctx.BlockHeight() {
		case v2.MainnetUpgradeHeight:
			upgradePlan.Name = v2.UpgradeName
			upgradePlan.Info = v2.UpgradeInfo
		case v4.MainnetUpgradeHeight:
			upgradePlan.Name = v4.UpgradeName
			upgradePlan.Info = v4.UpgradeInfo
		default:
			// No-op
			return
		}
	} else if strings.HasPrefix(ctx.ChainID(), TestnetChainID) {
		// handle testnet forks
		switch ctx.BlockHeight() {
		case v4.TestnetUpgradeHeight:
			upgradePlan.Name = v4.UpgradeName
			upgradePlan.Info = v4.UpgradeInfo
		default:
			// No-op
			return
		}
	} else {
		return
	}

	if err := app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan); err != nil {
		panic(
			fmt.Errorf(
				"failed to schedule upgrade %s during BeginBlock at height %d: %w",
				upgradePlan.Name, ctx.BlockHeight(), err,
			),
		)
	}
}
