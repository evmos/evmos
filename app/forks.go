// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ScheduleForkUpgrade executes any necessary fork logic for based upon the current
// block height and chain ID (mainnet or testnet). It sets an upgrade plan once
// the chain reaches the pre-defined upgrade height.
//
// CONTRACT: for this logic to work properly it is required to:
//
//  1. Release a non-breaking patch version so that the chain can set the scheduled upgrade plan at upgrade-height.
//  2. Release the software defined in the upgrade-info
func (app *Evmos) ScheduleForkUpgrade(_ sdk.Context) {
	// NOTE: there are no scheduled forks at the moment

	// upgradePlan := upgradetypes.Plan{
	//	Height: ctx.BlockHeight(),
	// }
	//
	// // handle mainnet forks with their corresponding upgrade name and info
	// switch ctx.BlockHeight() {
	// case v16.TestnetUpgradeHeight:
	//	upgradePlan.Name = v16.UpgradeNameTestnetRC4
	//	upgradePlan.Info = v16.UpgradeInfo
	// default:
	//	// No-op
	//	return
	// }
	//
	// // schedule the upgrade plan to the current block hight, effectively performing
	// // a hard fork that uses the upgrade handler to manage the migration.
	// if err := app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan); err != nil {
	//	panic(
	//		fmt.Errorf(
	//			"failed to schedule upgrade %s during BeginBlock at height %d: %w",
	//			upgradePlan.Name, ctx.BlockHeight(), err,
	//		),
	//	)
	// }
}
