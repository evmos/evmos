package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	claimskeeper "github.com/tharsis/evmos/v2/x/claims/keeper"
	erc20keeper "github.com/tharsis/evmos/v2/x/erc20/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v2
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	_ *erc20keeper.Keeper,
	_ *claimskeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// migrate claims and ERC20 module and don't perform custom logic here
		// Ref: https://docs.cosmos.network/master/building-modules/upgrade.html#registering-migrations
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
