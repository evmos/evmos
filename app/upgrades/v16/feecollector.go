package v16

import (
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// MigrateFeeCollector migrates the fee collector module account to include the `Burner` permission.
func MigrateFeeCollector(ak authkeeper.AccountKeeper, ctx sdk.Context, logger log.Logger) {
	feeCollectorModuleAccount := ak.GetModuleAccount(ctx, types.FeeCollectorName)
	if feeCollectorModuleAccount == nil {
		logger.Error("fee collector module account not found")
	}

	modAcc, ok := feeCollectorModuleAccount.(*types.ModuleAccount)
	if !ok {
		logger.Error("fee collector module account is not a module account")
	}

	// Create a new FeeCollector module account with the same address and balance as the old one.
	newFeeCollectorModuleAccount := types.NewModuleAccount(modAcc.BaseAccount, types.FeeCollectorName, types.Burner)

	// Override the FeeCollector module account in the auth keeper.
	ak.SetModuleAccount(ctx, newFeeCollectorModuleAccount)

}
