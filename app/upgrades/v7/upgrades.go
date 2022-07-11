package v7

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/evmos/evmos/v6/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v7
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bk bankkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// Refs:
		// - https://docs.cosmos.network/master/building-modules/upgrade.html#registering-migrations
		// - https://docs.cosmos.network/master/migrations/chain-upgrade-guide-044.html#chain-upgrade

		if types.IsTestnet(ctx.ChainID()) {
			logger.Debug("migrating inaccessible balance of secp faucet account...")
			MigrateFaucetBalance(ctx, bk)
		}

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// MigrateFaucetBalance transfers all balances of the inaccessible secp256k1
// Faucet address to a eth_secp256k1 address.
func MigrateFaucetBalance(ctx sdk.Context, bk bankkeeper.Keeper) {
	from := sdk.MustAccAddressFromBech32(FaucetAddressTo)
	to := sdk.MustAccAddressFromBech32(FaucetAddressTo)
	balances := bk.GetAllBalances(ctx, from)
	if err := bk.SendCoins(ctx, from, to, balances); err != nil {
		panic(err)
	}
}
