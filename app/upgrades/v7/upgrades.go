package v7

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/evoblockchain/evoblock/v8/types"
	claimskeeper "github.com/evoblockchain/evoblock/v8/x/claims/keeper"
	inflationkeeper "github.com/evoblockchain/evoblock/v8/x/inflation/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v7
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bk bankkeeper.Keeper,
	ik inflationkeeper.Keeper,
	ck *claimskeeper.Keeper,
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
		}

		if types.IsMainnet(ctx.ChainID()) {
			logger.Debug("migrating skipped epochs value of inflation module...")
			MigrateSkippedEpochs(ctx, ik)

			logger.Debug("migrating early contributor's claim record to new address...")
			MigrateContributorClaim(ctx, ck)
		}

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// MigrateFaucetBalances transfers all balances of the inaccessible secp256k1
// Faucet address on testnet to a eth_secp256k1 address.
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
// - launch date: 2022-03-02 20:00:00
// - halt date: 2022-03-06 22:11:42 (Block 58701)
// - relaunch date: 2022-04-27 18:00:00
// - inflation turned on: 2022-06-06 6:35:30 (skippedEpochs (incorrect) = 94 at this point)
// - counting mechanism fixed: 2022-07-04 (v6.0.0)
// = current date: 2022-07-07 13:00:00
// - currentEpochDay = 128
// = currentEpochWeek = 19
// - skippedEpochs (incorrectly calculated) = 94
// - 127 epochs have fully passed since launch
// - Of these 127 epochs, inflation has been enabled (at the end of the epoch) for 4 (pre-halt) + 31 (post-inflation-enabled) = 35 epochs
// - So the number of skippedEpochs (those with inflation disabled) should be 127 - 35 = 92 epochs, not 94 epochs
// - Can also see this by calculating number of completed epochs between halt date and date inflation turned on: 92 epochs between 3/6/2022 22:11:42, 6/6/2022 6:35:30
// Since skippedEpochs past v6.0.0 will be counted correctly (via PR #554), then we just account for the overcounting of 2 epochs
func MigrateSkippedEpochs(ctx sdk.Context, ik inflationkeeper.Keeper) {
	previousValue := ik.GetSkippedEpochs(ctx)
	newValue := previousValue - uint64(2)
	ik.SetSkippedEpochs(ctx, newValue)
}

// MigrateContributorClaim migrates the claims record of a specific early
// contributor from one address to another
func MigrateContributorClaim(ctx sdk.Context, k *claimskeeper.Keeper) {
	from, _ := sdk.AccAddressFromBech32(ContributorAddrFrom)
	to, _ := sdk.AccAddressFromBech32(ContributorAddrTo)

	cr, found := k.GetClaimsRecord(ctx, from)
	if !found {
		return
	}

	k.DeleteClaimsRecord(ctx, from)
	k.SetClaimsRecord(ctx, to, cr)
}
