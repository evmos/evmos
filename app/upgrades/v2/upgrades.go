package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const UpgradeName = "v2"

func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Set all modules "old versions" to 1.
		// Then the run migrations logic will handle running their upgrade logics
		// This will skip their InitGenesis
		fromVM := make(map[string]uint64)
		for moduleName := range mm.Modules {
			fromVM[moduleName] = 1
		}

		// TODO: Consensus versions for erc20 module and claims should be 2
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
