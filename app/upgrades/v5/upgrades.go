package v5

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	feestypes "github.com/tharsis/evmos/v5/x/fees/types"

	"github.com/tharsis/evmos/v5/types"
)

var TestnetDenomMetadata = banktypes.Metadata{
	Description: "The native EVM, governance and staking token of the Evmos testnet",
	DenomUnits: []*banktypes.DenomUnit{
		{
			Denom:    "atevmos",
			Exponent: 0,
			Aliases:  []string{"attotevmos"},
		},
		{
			Denom:    "tevmos",
			Exponent: 18,
		},
	},
	Base:    "atevmos",
	Display: "tevmos",
	Name:    "Testnet Evmos",
	Symbol:  "tEVMOS",
}

// CreateUpgradeHandler creates an SDK upgrade handler for v5
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bk bankkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Refs:
		// - https://docs.cosmos.network/master/building-modules/upgrade.html#registering-migrations
		// - https://docs.cosmos.network/master/migrations/chain-upgrade-guide-044.html#chain-upgrade

		// define the denom metadata for the testnet
		if types.IsTestnet(ctx.ChainID()) {
			bk.SetDenomMetaData(ctx, TestnetDenomMetadata)
		}

		// define from versions of the modules that have a new consensus version

		// migrate fee market module, other modules are left as-is to
		// avoid running InitGenesis.
		vm[feestypes.ModuleName] = 2

		// Leave modules are as-is to avoid running InitGenesis.
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
