// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v16

import (
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/evmos/evmos/v16/precompiles/bech32"
	osmosisoutpost "github.com/evmos/evmos/v16/precompiles/outposts/osmosis"
	strideoutpost "github.com/evmos/evmos/v16/precompiles/outposts/stride"
	"github.com/evmos/evmos/v16/precompiles/p256"
	"github.com/evmos/evmos/v16/utils"
	erc20keeper "github.com/evmos/evmos/v16/x/erc20/keeper"
	evmkeeper "github.com/evmos/evmos/v16/x/evm/keeper"
	inflationkeeper "github.com/evmos/evmos/v16/x/inflation/v1/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v16.0.0
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ek *evmkeeper.Keeper,
	bk bankkeeper.Keeper,
	inflationKeeper inflationkeeper.Keeper,
	ak authkeeper.AccountKeeper,
	gk govkeeper.Keeper,
	erck erc20keeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// Execute the conversion for all Cosmos native ERC20 token pairs to use the ERC20 EVM extension.
		if err := ConvertToNativeCoinExtensions(ctx, logger, ak, bk, erck); err != nil {
			logger.Error("failed to convert erc20s to native coins", "error", err.Error())
		}

		// enable secp256r1 and bech32 precompile on testnet
		if utils.IsTestnet(ctx.ChainID()) {
			p256Address := p256.Precompile{}.Address()
			bech32Address := bech32.Precompile{}.Address()
			if err := ek.EnablePrecompiles(ctx, p256Address, bech32Address); err != nil {
				logger.Error("failed to enable precompiles", "error", err.Error())
			}
		}

		// enable stride and osmosis outposts
		strideAddress := strideoutpost.Precompile{}.Address()
		osmosisAddress := osmosisoutpost.Precompile{}.Address()
		if err := ek.EnablePrecompiles(ctx, strideAddress, osmosisAddress); err != nil {
			logger.Error("failed to enable outposts", "error", err.Error())
		}

		// Migrate the FeeCollector module account to include the Burner permission.
		// This is required when including the postHandler to burn Cosmos Tx fees
		if err := MigrateFeeCollector(ak, ctx); err != nil {
			logger.Error("failed to migrate the fee collector", "error", err.Error())
		}

		// TODO: uncomment when ready
		// if err := BurnUsageIncentivesPool(ctx, bankKeeper); err != nil {
		//	logger.Error("failed to burn inflation pool", "error", err.Error())
		// }

		if err := UpdateInflationParams(ctx, inflationKeeper); err != nil {
			logger.Error("failed to update inflation params", "error", err.Error())
		}

		// Remove the deprecated governance proposals from store
		logger.Debug("deleting deprecated proposals...")
		DeleteDeprecatedProposals(ctx, gk, logger)

		// recovery module is deprecated
		logger.Debug("deleting recovery module from version map...")
		delete(vm, "recovery")
		logger.Debug("deleting claims module from version map...")
		delete(vm, "claims")
		logger.Debug("deleting incentives module from version map...")
		delete(vm, "incentives")

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// ConvertToNativeCoinExtensions converts all the registered ERC20 tokens of Cosmos native token pairs
// back to the native representation and registers the (W)ERC20 precompiles for each token pair.
func ConvertToNativeCoinExtensions(
	ctx sdk.Context,
	logger log.Logger,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
) error {
	// NOTE (@fedekunze): first we must convert the all the registered tokens.
	// If we do it the other way around, the conversion will fail since there won't
	// be any contract code due to the selfdestruct.
	if err := ConvertERC20Coins(ctx, logger, accountKeeper, bankKeeper, erc20Keeper); err != nil {
		logger.Error("failed to convert native coins", "error", err.Error())
		// TODO: return error here?
		// return errorsmod.Wrap(err, "failed to convert native coins")
	}

	// Instantiate the (W)ERC20 Precompile for each registered IBC Coin

	// IMPORTANT (@fedekunze): This logic needs to be included on EVERY UPGRADE
	// from now on because the AvailablePrecompiles function does not have access
	// to the state (in this case, the registered token pairs).
	//
	// FIXME: Do we want to run both if the convert coins failed? I think that could be dangerous if a coin
	// was not fully converted and we overwrite the contract address with the Extension? Or maybe not?
	if err := erc20Keeper.RegisterERC20Extensions(ctx); err != nil {
		logger.Error("failed to register ERC-20 Extensions", "error", err.Error())
		// TODO: return error here?
		// return errorsmod.Wrap(err, "failed to convert native coins")
	}

	return nil
}
