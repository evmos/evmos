package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	erc20keeper "github.com/tharsis/evmos/x/erc20/keeper"
)

const UpgradeName = "v2"

func CreateUpgradeHandler(mm *module.Manager, configurator module.Configurator,
	erc20Keeper *erc20keeper.Keeper,
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
		params := erc20Keeper.GetParams(ctx)
		params.EnableEVMHook = true
		params.EnableErc20 = true
		erc20Keeper.SetParams(ctx, params)

		// Claims
		// Two parameters added, how do we set them?
		// copy the struct on genesis.pb.go file

		// TODO: Consensus versions for claims should be 2
		// But do we need to bump the consensus versions for the ERC20 module?
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
