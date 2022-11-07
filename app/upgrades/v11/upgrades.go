package v11

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	erc20keeper "github.com/evmos/evmos/v10/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v10/x/erc20/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v11
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	ak authkeeper.AccountKeeper,
	bk bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		cacheCtx, writeFn := ctx.CacheContext()
		if err := ConvertRegisteredCoins(cacheCtx, ak, bk, erc20Keeper); err != nil {
			logger.Error("failed to convert registered coins", "error", err.Error())
		} else {
			writeFn()
		}

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// ConvertRegisteredCoins converts all the registered coins to their corresponding ERC20 tokens.
func ConvertRegisteredCoins(ctx sdk.Context,
	ak authkeeper.AccountKeeper,
	bk bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
) (err error) {
	if !erc20Keeper.IsERC20Enabled(ctx) {
		return nil
	}

	registeredIBCVouchers := make(map[string]bool)

	// iterate over registered token pairs and add them to the map
	erc20Keeper.IterateTokenPairs(ctx, func(tokenPair erc20types.TokenPair) (stop bool) {
		// only register IBC vouchers
		if strings.HasPrefix(tokenPair.Denom, "ibc/") {
			registeredIBCVouchers[tokenPair.Denom] = true
		}

		return false
	})

	// iterate over balances and convert the IBC voucher coins to ERC20s
	bk.IterateAllBalances(ctx, func(address sdk.AccAddress, coin sdk.Coin) (stop bool) {
		if !registeredIBCVouchers[coin.Denom] {
			return false
		}

		acc := ak.GetAccount(ctx, address)

		// don't convert balances from module accounts
		if _, isModuleAccount := acc.(authtypes.ModuleAccountI); acc == nil || isModuleAccount {
			return false
		}

		addr := address.String()

		msg := &erc20types.MsgConvertCoin{
			Coin:     coin,
			Sender:   addr,
			Receiver: addr,
		}

		// convert coin
		_, err = erc20Keeper.ConvertCoin(sdk.WrapSDKContext(ctx), msg)
		// stop iteration on error
		return err != nil
	})

	// return error from iterator
	if err != nil {
		return err
	}

	return nil
}
