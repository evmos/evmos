package v4

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibcclientkeeper "github.com/cosmos/ibc-go/v3/modules/core/02-client/keeper"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"

	feemarketkeeper "github.com/tharsis/ethermint/x/feemarket/keeper"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	claimskeeper "github.com/tharsis/evmos/v3/x/claims/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v2
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	clientKeeper ibcclientkeeper.Keeper,
	slashingKeeper slashingkeeper.Keeper,
	feemarketKeeper feemarketkeeper.Keeper,
	claimsKeeper *claimskeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Refs:
		// - https://docs.cosmos.network/master/building-modules/upgrade.html#registering-migrations
		// - https://docs.cosmos.network/master/migrations/chain-upgrade-guide-044.html#chain-upgrade

		updateSlashingParams(ctx, slashingKeeper)
		if err := updateIBCClients(ctx, clientKeeper); err != nil {
			return vm, err
		}

		updateFeeMarketParams(ctx, feemarketKeeper)
		updateClaimsParams(ctx, claimsKeeper)

		// Leave modules are as-is to avoid running InitGenesis.
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// updateSlashingParams changes the slashing window from 30k to 10k
func updateSlashingParams(ctx sdk.Context, k slashingkeeper.Keeper) {
	params := k.GetParams(ctx)
	params.SignedBlocksWindow = 10_000
	k.SetParams(ctx, params)
}

func updateIBCClients(ctx sdk.Context, k ibcclientkeeper.Keeper) error {
	proposalOsmosis := &ibcclienttypes.ClientUpdateProposal{
		Title:              "Update expired Osmosis IBC client",
		Description:        "Update existing Cosmos Hub IBC client on Evmos (07-tendermint-0) in order to resume packet transfers between both chains.",
		SubjectClientId:    "07-tendermint-0", // Osmosis
		SubstituteClientId: "07-tendermint-1", // TODO: replace with new client
	}

	proposalCosmosHub := &ibcclienttypes.ClientUpdateProposal{
		Title:              "Update expired Cosmos Hub IBC client",
		Description:        "Update existing Cosmos Hub IBC client on Evmos (07-tendermint-3) in order to resume packet transfers between both chains.",
		SubjectClientId:    "07-tendermint-3", // Cosmos Hub
		SubstituteClientId: "07-tendermint-1", // TODO: replace with new client
	}

	if err := k.ClientUpdateProposal(ctx, proposalOsmosis); err != nil {
		return sdkerrors.Wrap(err, "failed to update Osmosis IBC client")
	}

	if err := k.ClientUpdateProposal(ctx, proposalCosmosHub); err != nil {
		return sdkerrors.Wrap(err, "failed to update Cosmos Hub IBC client")
	}

	return nil
}

func updateFeeMarketParams(ctx sdk.Context, k feemarketkeeper.Keeper) {
	params := k.GetParams(ctx)
	params.BaseFee = feemarkettypes.DefaultParams().BaseFee // 1 billion
	params.ElasticityMultiplier = 4
	k.SetParams(ctx, params)
}

func updateClaimsParams(ctx sdk.Context, k *claimskeeper.Keeper) {
	params := k.GetParams(ctx)
	params.DurationUntilDecay += time.Hour * 24 * 14 // add 14 days
	k.SetParams(ctx, params)
}
