package v9

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	distrKeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/evmos/evmos/v9/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v9
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	dk distrKeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		if types.IsMainnet(ctx.ChainID()) {
			logger.Debug("recovering lost funds from clawback...")
			ReturnFundsFromCommunityPool(ctx, dk)
		}

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func ReturnFundsFromCommunityPool(ctx sdk.Context, dk distrKeeper.Keeper) {
	for i := range Accounts {
		if err := ReturnFundsFromCommunityPoolToAccount(ctx, dk, Accounts[i][0], Accounts[i][1]); err != nil {
			panic(err) // TODO: CHECK WHAT TO DO IN THIS CASE
		}
	}
}

func ReturnFundsFromCommunityPoolToAccount(ctx sdk.Context, dk distrKeeper.Keeper, account string, amount string) error {
	to := sdk.MustAccAddressFromBech32(account)
	res, _ := sdkmath.NewIntFromString(amount)
	balance := sdk.NewCoin("aevmos", res)

	if err := dk.DistributeFromFeePool(ctx, sdk.NewCoins(balance), to); err != nil {
		return err
	}
	return nil
}
