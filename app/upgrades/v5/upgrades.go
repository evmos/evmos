package v5

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	epochstypes "github.com/tharsis/evmos/v4/x/epochs/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v5
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Refs:
		// - https://docs.cosmos.network/master/building-modules/upgrade.html#registering-migrations
		// - https://docs.cosmos.network/master/migrations/chain-upgrade-guide-044.html#chain-upgrade

		// migrate epochs module, other modules are left as-is to
		// avoid running InitGenesis.
		vm[epochstypes.ModuleName] = 1

		// Leave modules are as-is to avoid running InitGenesis.
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
