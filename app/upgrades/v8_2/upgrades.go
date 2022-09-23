package v82

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	revenuekeeper "github.com/evmos/evmos/v8/x/revenue/keeper"
	revenuetypes "github.com/evmos/evmos/v8/x/revenue/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v8.2
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	rk revenuekeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		logger.Debug("setting parameters to default parameters in revenue module...")
		SetRevenueParameters(ctx, rk)

		// Refs:
		// - https://docs.cosmos.network/master/building-modules/upgrade.html#registering-migrations
		// - https://docs.cosmos.network/master/migrations/chain-upgrade-guide-044.html#chain-upgrade

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func SetRevenueParameters(ctx sdk.Context, rk revenuekeeper.Keeper) {
	rk.SetParams(ctx, revenuetypes.DefaultParams())
}
