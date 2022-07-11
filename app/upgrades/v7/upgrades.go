package v7

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/evmos/evmos/v6/types"
	inflationkeeper "github.com/evmos/evmos/v6/x/inflation/keeper"
	inflationtypes "github.com/evmos/evmos/v6/x/inflation/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v7
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bk bankkeeper.Keeper,
	ik inflationkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// Refs:
		// - https://docs.cosmos.network/master/building-modules/upgrade.html#registering-migrations
		// - https://docs.cosmos.network/master/migrations/chain-upgrade-guide-044.html#chain-upgrade

		if types.IsTestnet(ctx.ChainID()) {
			logger.Debug("migrating inaccessible balance of secp faucet account...")
			if err := MigrateFaucetBalances(ctx, bk); err != nil {
				// log error instead of aborting the upgrade
				logger.Error("FAILED TO MIGRATE FAUCET BALANCES", "error", err.Error())
			}

			logger.Debug("migrating skipped epochs value of inflation module...")
			MigrateSkippedEpochs(ctx, ik)
		}

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// MigrateFaucetBalances transfers all balances of the inaccessible secp256k1
// Faucet address to a eth_secp256k1 address.
func MigrateFaucetBalances(ctx sdk.Context, bk bankkeeper.Keeper) error {
	from := sdk.MustAccAddressFromBech32(FaucetAddressFrom)
	to := sdk.MustAccAddressFromBech32(FaucetAddressTo)
	balances := bk.GetAllBalances(ctx, from)
	if err := bk.SendCoins(ctx, from, to, balances); err != nil {
		return sdkerrors.Wrap(err, "failed to migrate Faucet Balances")
	}
	return nil
}

// MigrateSkippedEpochs migrates the number of skipped epochs to be lower
// than the previous stored value, due to an overcounting of two epochs pre v6.0.0.
func MigrateSkippedEpochs(sdk sdk.Context, ik inflationkeeper.Keeper) error {
	previousValue := ik.GetSkippedEpochs(ctx)
	newValue := previousValue - uint64(2)
	ik.SetSkippedEpochs(ctx, newValue)
	return nil
}
