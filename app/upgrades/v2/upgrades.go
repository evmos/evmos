package v2

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	claimstypes "github.com/tharsis/evmos/x/claims/types"

	claimskeeper "github.com/tharsis/evmos/x/claims/keeper"
	erc20keeper "github.com/tharsis/evmos/x/erc20/keeper"
)

func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator,
	erc20Keeper *erc20keeper.Keeper,
	claimsKeeper *claimskeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Set all modules "old versions" to 1.
		// Then the run migrations logic will handle running their upgrade logics
		// This will skip their InitGenesis
		fromVM := make(map[string]uint64)
		for moduleName := range mm.Modules {
			fromVM[moduleName] = 1
		}

		// Set the params for the erc20 module
		erc20Params := erc20Keeper.GetParams(ctx)
		erc20Params.EnableEVMHook = true
		erc20Params.EnableErc20 = true
		erc20Keeper.SetParams(ctx, erc20Params)

		// Claims
		claimsParams := claimsKeeper.GetParams(ctx)
		claimsParams.DurationUntilDecay += time.Hour * 24 * 14 // add 2 weeks
		claimsParams.AuthorizedChannels = claimstypes.DefaultAuthorizedChannels
		claimsParams.EVMChannels = claimstypes.DefaultEVMChannels
		claimsKeeper.SetParams(ctx, claimsParams)

		// Bump the consensus version for claims so that InitGenesis will run
		fromVM[claimstypes.ModuleName] = 2

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
