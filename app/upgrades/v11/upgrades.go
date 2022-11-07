package v11

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	erc20keeper "github.com/evmos/evmos/v10/x/erc20/keeper"
	erc20types "github.com/evmos/evmos/v10/x/erc20/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v10
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bk bankkeeper.Keeper,
	erc20Keeper erc20keeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		if err := ConvertRegisteredCoins(ctx, bk, erc20Keeper); err != nil {
			return nil, err
		}

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// ConvertRegisteredCoins converts all the registered coins to their corresponding ERC20 tokens.
func ConvertRegisteredCoins(ctx sdk.Context, bk bankkeeper.Keeper, erc20Keeper erc20keeper.Keeper) (err error) {
	params := erc20Keeper.GetParams(ctx)
	if !params.EnableErc20 {
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
