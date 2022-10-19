package v9

import (
	"fmt"

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
			if err := ReturnFundsFromCommunityPool(ctx, dk); err != nil {
				// log error instead of aborting the upgrade
				logger.Error("FAILED TO RECOVER FROM COMMUNITY FUNDS", "error", err.Error())
			}
		}

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func ReturnFundsFromCommunityPool(ctx sdk.Context, dk distrKeeper.Keeper) error {
	availableCoins, _ := sdkmath.NewIntFromString(MaxRecover)
	for i := range Accounts {
		refund, _ := sdkmath.NewIntFromString(Accounts[i][1])
		if availableCoins.LT(refund) {
			return fmt.Errorf("refund exceeds the total available coins: %s > %s", Accounts[i][1], availableCoins)
		}
		if err := ReturnFundsFromCommunityPoolToAccount(ctx, dk, Accounts[i][0], refund); err != nil {
			return err
		}
		availableCoins = availableCoins.Sub(transferCoin)
	}
	return nil
}

func ReturnFundsFromCommunityPoolToAccount(ctx sdk.Context, dk distrKeeper.Keeper, account string, amount string) error {
	to := sdk.MustAccAddressFromBech32(account)
	res, _ := sdk.NewIntFromString(amount)
	balance := sdk.Coin{
		Denom: "aevmos",
		Amount: res,
	}

	if err := dk.DistributeFromFeePool(ctx, sdk.Coins{balance}, to); err != nil {
		return err
	}
	return nil
}
