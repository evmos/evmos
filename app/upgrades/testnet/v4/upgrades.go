package tv4

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	feemarketkeeper "github.com/tharsis/ethermint/x/feemarket/keeper"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	claimskeeper "github.com/tharsis/evmos/v3/x/claims/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v4. It handles the upgrade
// functionality by importing the exported state. It doesn't perform any in-place store
// migrations for the modules since there are no consensus version bumps. Aditionally,
// it updates the slashing, fee market and claims params.

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	slashingKeeper slashingkeeper.Keeper,
	feemarketKeeper feemarketkeeper.Keeper,
	claimsKeeper *claimskeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Refs:
		// - https://docs.cosmos.network/master/building-modules/upgrade.html#registering-migrations
		// - https://docs.cosmos.network/master/migrations/chain-upgrade-guide-044.html#chain-upgrade

		updateSlashingParams(ctx, slashingKeeper)
		updateFeeMarketParams(ctx, feemarketKeeper)
		updateClaimsParams(ctx, claimsKeeper)

		// Leave modules are as-is to avoid running InitGenesis.
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// updateSlashingParams changes the slashing window from 30k to 10k
func updateSlashingParams(ctx sdk.Context, k slashingkeeper.Keeper) {
	params := k.GetParams(ctx)
	params.SignedBlocksWindow = 10_000
	k.SetParams(ctx, params)
}

// updateFeeMarketParams changes the base fee param to the default (1 billion)
// and also bumps the elasticity multiplier to 4
func updateFeeMarketParams(ctx sdk.Context, k feemarketkeeper.Keeper) {
	params := k.GetParams(ctx)
	params.BaseFee = feemarkettypes.DefaultParams().BaseFee // 1 billion
	params.ElasticityMultiplier = 4
	k.SetParams(ctx, params)
}

// updateClaimsParams adds 14 days to the claim duration
func updateClaimsParams(ctx sdk.Context, k *claimskeeper.Keeper) {
	params := k.GetParams(ctx)
	params.DurationUntilDecay += time.Hour * 24 * 14 // add 14 days
	k.SetParams(ctx, params)
}
