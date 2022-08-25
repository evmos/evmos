package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	claimstypes "github.com/evmos/evmos/v9/x/claims/types"
	erc20types "github.com/evmos/evmos/v9/x/erc20/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v2
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Refs:
		// - https://docs.cosmos.network/master/building-modules/upgrade.html#registering-migrations
		// - https://docs.cosmos.network/master/migrations/chain-upgrade-guide-044.html#chain-upgrade

		// migrate claims and ERC20 module, other modules are left as-is to
		// avoid running InitGenesis.
		vm[claimstypes.ModuleName] = 1
		vm[erc20types.ModuleName] = 1

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
