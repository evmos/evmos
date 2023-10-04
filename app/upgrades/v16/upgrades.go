// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package v16

import (
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v14/contracts"
	evmostypes "github.com/evmos/evmos/v14/types"
	erc20keeper "github.com/evmos/evmos/v14/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v14/x/erc20/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v16
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// NOTE (@fedekunze): first we must convert the all the registered tokens.
		// If we do it the other way around, the conversion will fail since there won't
		// be any contract code due to the selfdestruct.
		if err := ConvertNativeCoins(ctx, logger, accountKeeper, bankKeeper, erc20Keeper); err != nil {
			logger.Error("failed to convert native coins", "error", err.Error())
		}

		// Instantiate the (W)ERC20 Precompile for each registered IBC Coin

		// IMPORTANT (@fedekunze): This logic needs to be included on EVERY UPGRADE
		// from now on because the AvailablePrecompiles function does not have access
		// to the state (in this case, the registered token pairs).
		if err := erc20Keeper.RegisterERC20Extensions(ctx); err != nil {
			logger.Error("failed to register ERC-20 Extensions", "error", err.Error())
		}

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func ConvertNativeCoins(
	ctx sdk.Context,
	logger log.Logger,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
) error {
	// iterate over all the accounts and convert the tokens to native coins
	accountKeeper.IterateAccounts(ctx, func(account authtypes.AccountI) (stop bool) {
		ethAccount, ok := account.(*evmostypes.EthAccount)
		if !ok {
			return false
		}

		// TODO: convert WEVMOS to EVMOS by using "withdraw" method

		erc20Keeper.IterateTokenPairs(ctx, func(tokenPair erc20types.TokenPair) bool {
			if !tokenPair.IsNativeCoin() {
				return false
			}

			ethAddress := ethAccount.EthAddress()

			balance := erc20Keeper.BalanceOf(ctx, contracts.ERC20MinterBurnerDecimalsContract.ABI, tokenPair.GetERC20Contract(), ethAddress)
			if balance.Cmp(common.Big0) <= 0 {
				return false
			}

			cosmosAddress := sdk.AccAddress(ethAddress.Bytes())

			msg := erc20types.NewMsgConvertCoin(sdk.Coin{Denom: tokenPair.Denom, Amount: sdk.NewIntFromBigInt(balance)}, ethAddress, cosmosAddress)

			// TODO: use the legacy logic here to burn the ERC20s and unlock the native tokens
			_, err := erc20Keeper.ConvertCoin(sdk.WrapSDKContext(ctx), msg)
			if err != nil {
				logger.Debug(
					"failed to convert coin",
					"account", cosmosAddress.String(),
					"coin", tokenPair.Denom,
					"balance", balance.String(),
					"error", err.Error())
			}

			return false
		})

		return false
	})

	erc20ModuleAccountAddress := authtypes.NewModuleAddress(erc20types.ModuleName)
	balances := bankKeeper.GetAllBalances(ctx, erc20ModuleAccountAddress)
	if balances.IsZero() {
		return nil
	}

	// burn all the coins left in the module account
	return bankKeeper.BurnCoins(ctx, erc20types.ModuleName, balances)
}
