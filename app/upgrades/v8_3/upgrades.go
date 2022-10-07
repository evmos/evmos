package v83

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	distrKeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/evmos/evmos/v9/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v8.3
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	dk distrKeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// if types.IsTestnet(ctx.ChainID()) {
		// 	logger.Debug("migrate feesplit to revenue...")
		// }

		if types.IsMainnet(ctx.ChainID()) {
			logger.Debug("recovering lost funds from clawback...")
			for account, fund := range accounts {
				if err := returnFundsFromCommunityPool(ctx, dk, account, fund); err != nil {
					panic(err) // TODO: CHECK WHAT TO DO IN THIS CASE
				}
			}
		}

		if types.IsMainnet(ctx.ChainID()) {
			logger.Debug("migrate IBC module account")
		}

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func returnFundsFromCommunityPool(ctx sdk.Context, dk distrKeeper.Keeper, account string, amount string) error {
	to := sdk.MustAccAddressFromBech32(account)
	res, _ := sdkmath.NewIntFromString(amount)
	balance := sdk.NewCoin("aevmos", res)

	if err := dk.DistributeFromFeePool(ctx, sdk.NewCoins(balance), to); err != nil {
		return err
	}
	return nil
}
