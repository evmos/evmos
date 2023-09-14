// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v14rc5

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	distprecompile "github.com/evmos/evmos/v14/precompiles/distribution"
	ics20precompile "github.com/evmos/evmos/v14/precompiles/ics20"
	stakingprecompile "github.com/evmos/evmos/v14/precompiles/staking"
	evmkeeper "github.com/evmos/evmos/v14/x/evm/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v13
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ek *evmkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// Disable vesting precompile in the EVM module
		//
		// NOTE: This only serves testing purposes and should NOT be used in the mainnet handler.
		// In order to test the changed behavior of calling non-active EVM extensions, it is necessary
		// to disable one to test this.
		evmParams := ek.GetParams(ctx)
		evmParams.ActivePrecompiles = []string{
			stakingprecompile.Precompile{}.Address().String(),
			distprecompile.Precompile{}.Address().String(),
			ics20precompile.Precompile{}.Address().String(),
		}
		err := ek.SetParams(ctx, evmParams)
		if err != nil {
			// log error instead of aborting the upgrade
			logger.Error("failed to set EVM params", "err", err)
		}

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
