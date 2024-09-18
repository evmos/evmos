// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v20

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	evmkeeper "github.com/evmos/evmos/v20/x/evm/keeper"
	"github.com/evmos/evmos/v20/x/evm/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for Evmos v20
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ek *evmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		logger := ctx.Logger().With("upgrade", UpgradeName)

		logger.Debug("Enabling gov precompile...")
		if err := EnableGovPrecompile(ctx, ek); err != nil {
			logger.Error("error while enabling gov precompile", "error", err.Error())
		}

		logger.Debug("Running module migrations...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func EnableGovPrecompile(ctx sdk.Context, ek *evmkeeper.Keeper) error {
	// Enable gov precompile
	params := ek.GetParams(ctx)
	params.ActiveStaticPrecompiles = append(params.ActiveStaticPrecompiles, types.GovPrecompileAddress)
	if err := params.Validate(); err != nil {
		return err
	}
	return ek.SetParams(ctx, params)
}
