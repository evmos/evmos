// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v17

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/ethereum/go-ethereum/common"
	erc20precompile "github.com/evmos/evmos/v16/precompiles/erc20"
	"github.com/evmos/evmos/v16/utils"
	erc20keeper "github.com/evmos/evmos/v16/x/erc20/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v17.0.0
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ak authkeeper.AccountKeeper,
	bk bankkeeper.Keeper,
	erck erc20keeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		var wrappedContractAddr common.Address
		switch {
		case utils.IsMainnet(ctx.ChainID()):
			wrappedContractAddr = common.HexToAddress(erc20precompile.WEVMOSContractMainnet)
		case utils.IsTestnet(ctx.ChainID()):
			wrappedContractAddr = common.HexToAddress(erc20precompile.WEVMOSContractTestnet)
		default:
			logger.Error("unexpected chain id", "chain-id", ctx.ChainID())
		}

		// Execute the conversion for all Cosmos native ERC20 token pairs to use the ERC20 EVM extension.
		cacheCtx, writeFn := ctx.CacheContext()
		if err := ConvertToNativeCoinExtensions(cacheCtx, logger, ak, bk, erck, wrappedContractAddr); err != nil {
			logger.Error("failed to fully convert erc20s to native coins", "error", err.Error())
		} else {
			// Write the cache to the context
			writeFn()
		}

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
