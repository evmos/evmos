// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package app

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	v17 "github.com/evmos/evmos/v18/app/upgrades/v17"
	"github.com/evmos/evmos/v18/utils"
)

// ScheduleForkUpgrade executes any necessary fork logic for based upon the current
// block height and chain ID (mainnet or testnet). It sets an upgrade plan once
// the chain reaches the pre-defined upgrade height.
//
// CONTRACT: for this logic to work properly it is required to:
//
//  1. Release a non-breaking patch version so that the chain can set the scheduled upgrade plan at upgrade-height.
//  2. Release the software defined in the upgrade-info
func (app *Evmos) ScheduleForkUpgrade(ctx sdk.Context) {
	// Only fork on Mainnet
	if !utils.IsMainnet(ctx.ChainID()) {
		return
	}
	upgradePlan := upgradetypes.Plan{
		Height: ctx.BlockHeight(),
	}

	// handle mainnet forks with their corresponding upgrade name and info
	switch ctx.BlockHeight() {
	case v17.MainnetUpgradeHeight:
		upgradePlan.Name = v17.UpgradeName
		upgradePlan.Info = v17.UpgradeInfo
	default:
		// No-op
		return
	}

	// schedule the upgrade plan to the current block height, effectively performing
	// a hard fork that uses the upgrade handler to manage the migration.
	if err := app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan); err != nil {
		panic(
			fmt.Errorf(
				"failed to schedule upgrade %s during BeginBlock at height %d: %w",
				upgradePlan.Name, ctx.BlockHeight(), err,
			),
		)
	}
}
