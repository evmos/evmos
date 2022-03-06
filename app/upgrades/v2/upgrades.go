package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	claimstypes "github.com/tharsis/evmos/v2/x/claims/types"

	claimskeeper "github.com/tharsis/evmos/v2/x/claims/keeper"
	erc20keeper "github.com/tharsis/evmos/v2/x/erc20/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v2
func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator,
	erc20Keeper *erc20keeper.Keeper,
	claimsKeeper *claimskeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Set the params for the erc20 module here as we are not bumping consensus breaking
		erc20Params := erc20Keeper.GetParams(ctx)
		erc20Params.EnableEVMHook = true
		erc20Params.EnableErc20 = true
		erc20Keeper.SetParams(ctx, erc20Params)

		// Set the consensus version to the from version for claims so that InitGenesis doesn't run
		// with a default empty claim's GenesisState and the migration code is executed.
		vm[claimstypes.ModuleName] = 1

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
