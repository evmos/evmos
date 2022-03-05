package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func SetupUpgradeHandlers(app *Evmos) {
	app.UpgradeKeeper.SetUpgradeHandler("v2.0.0", func(ctx sdk.Context, plan upgradetypes.Plan, _ module.VersionMap) (module.VersionMap, error) {

		// Set all modules "old versions" to 1.
		// Then the run migrations logic will handle running their upgrade logics
		// This will skip their InitGenesis
		fromVM := make(map[string]uint64)
		for moduleName := range app.mm.Modules {
			fromVM[moduleName] = 1
		}

		// TODO: Consensus versions for erc20 module and claims should be 2
		return app.mm.RunMigrations(ctx, app.configurator, fromVM)
	})
}
