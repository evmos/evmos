// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package v91

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	distrKeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/evmos/evmos/v12/utils"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v9
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	dk distrKeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		if utils.IsMainnet(ctx.ChainID()) {
			logger.Debug("recovering lost funds from clawback...")
			HandleMainnetUpgrade(ctx, dk, logger)
		}

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// HandleMainnetUpgrade handles the logic for Mainnet upgrade, it only commits to the db if successful
func HandleMainnetUpgrade(ctx sdk.Context, dk distrKeeper.Keeper, logger log.Logger) {
	// use a cache context as a rollback mechanism in case
	// the refund fails
	cacheCtx, writeFn := ctx.CacheContext()
	err := ReturnFundsFromCommunityPool(cacheCtx, dk)
	if err != nil {
		// log error instead of aborting the upgrade
		logger.Error("failed to recover from community funds", "error", err.Error())
	} else {
		writeFn()
	}
}

// ReturnFundsFromCommunityPool handles the return of funds from the community pool to accounts affected during the claims clawback
func ReturnFundsFromCommunityPool(ctx sdk.Context, dk distrKeeper.Keeper) error {
	availableCoins, ok := sdk.NewIntFromString(MaxRecover)
	if !ok || availableCoins.IsNegative() {
		return fmt.Errorf("failed to read maximum amount to recover from community funds")
	}
	for i := range Accounts {
		address := Accounts[i][0]
		amt := Accounts[i][1]

		refund, _ := sdk.NewIntFromString(amt)
		if availableCoins.LT(refund) {
			return fmt.Errorf(
				"refund to address %s exceeds the total available coins: %s > %s",
				address, amt, availableCoins,
			)
		}
		if err := ReturnFundsFromCommunityPoolToAccount(ctx, dk, address, refund); err != nil {
			return err
		}
		availableCoins = availableCoins.Sub(refund)
	}
	return nil
}

// ReturnFundsFromCommunityPoolToAccount sends specified amount from the community pool to the affected account
func ReturnFundsFromCommunityPoolToAccount(ctx sdk.Context, dk distrKeeper.Keeper, account string, amount sdkmath.Int) error {
	to := sdk.MustAccAddressFromBech32(account)
	balance := sdk.Coin{
		Denom:  "aevmos",
		Amount: amount,
	}
	return dk.DistributeFromFeePool(ctx, sdk.Coins{balance}, to)
}
