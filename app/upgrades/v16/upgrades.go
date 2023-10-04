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
	"github.com/evmos/evmos/v14/utils"
	erc20keeper "github.com/evmos/evmos/v14/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v14/x/erc20/types"
)

const (
	WEVMOSContractMainnet = "0xD4949664cD82660AaE99bEdc034a0deA8A0bd517"
	WEVMOSContractTestnet = ""
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
	var wrappedContractAddr common.Address
	isMainnet := utils.IsMainnet(ctx.ChainID())
	if isMainnet {
		wrappedContractAddr = common.HexToAddress(WEVMOSContractMainnet)
	}

	// iterate over all the accounts and convert the tokens to native coins
	accountKeeper.IterateAccounts(ctx, func(account authtypes.AccountI) (stop bool) {
		ethAccount, ok := account.(*evmostypes.EthAccount)
		if !ok {
			return false
		}

		ethAddress := ethAccount.EthAddress()
		cosmosAddress := sdk.AccAddress(ethAddress.Bytes())

		// TODO: convert WEVMOS to EVMOS by using "withdraw" method

		if isMainnet {
			balance := erc20Keeper.BalanceOf(ctx, contracts.ERC20MinterBurnerDecimalsContract.ABI, wrappedContractAddr, ethAddress)

			// only execute the withdrawal if balance is positive
			if balance.Cmp(common.Big0) > 0 {
				// call withdraw method from the account
				data := []byte{}
				res, err := erc20Keeper.CallEVMWithData(ctx, ethAddress, &wrappedContractAddr, data, true)
				if err != nil {
					logger.Debug(
						"failed to withdraw WEVMOS",
						"account", cosmosAddress.String(),
						"balance", balance.String(),
						"error", err.Error(),
					)
				} else if res.VmError != "" {
					logger.Debug(
						"withdraw WEVMOS reverted",
						"account", cosmosAddress.String(),
						"balance", balance.String(),
						"vm-error", res.VmError,
					)
				}
			}
		}

		erc20Keeper.IterateTokenPairs(ctx, func(tokenPair erc20types.TokenPair) bool {
			if !tokenPair.IsNativeCoin() {
				return false
			}

			contract := tokenPair.GetERC20Contract()

			balance := erc20Keeper.BalanceOf(ctx, contracts.ERC20MinterBurnerDecimalsContract.ABI, contract, ethAddress)
			if balance.Cmp(common.Big0) <= 0 {
				return false
			}

			msg := erc20types.NewMsgConvertERC20(sdk.NewIntFromBigInt(balance), cosmosAddress, contract, ethAddress)

			_, err := erc20Keeper.ConvertERC20(sdk.WrapSDKContext(ctx), msg)
			if err != nil {
				logger.Debug(
					"failed to convert ERC20 to native Coin",
					"account", ethAddress.String(),
					"erc20", contract.String(),
					"balance", balance.String(),
					"error", err.Error(),
				)
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
