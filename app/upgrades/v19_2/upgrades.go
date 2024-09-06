// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v192

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	erc20keeper "github.com/evmos/evmos/v19/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v19/x/erc20/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v19.2
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	erc20k erc20keeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		logger := ctx.Logger().With("upgrade", UpgradeName)

		if err := AddCodeToERC20Extensions(ctx, logger, erc20k); err == nil {
			return nil, err
		}

		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// AddCodeToERC20Extensions adds code and code hash to the ERC20 precompiles with the EVM.
func AddCodeToERC20Extensions(
	ctx sdk.Context,
	logger log.Logger,
	erc20Keeper erc20keeper.Keeper,
) (err error) {
	logger.Info("Adding code to erc20 extensions...")

	erc20Keeper.IterateTokenPairs(ctx, func(tokenPair erc20types.TokenPair) bool {
		// Only need to add code to the IBC coins.
		if !tokenPair.IsNativeCoin() {
			return false
		}

		err = erc20Keeper.RegisterERC20CodeHash(ctx, tokenPair.GetERC20Contract())
		return err != nil
	})

	logger.Info("Done with erc20 extensions")
	return err
}
