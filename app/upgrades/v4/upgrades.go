package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibcclientkeeper "github.com/cosmos/ibc-go/v3/modules/core/02-client/keeper"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"

	"github.com/tharsis/evmos/v5/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v4
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	clientKeeper ibcclientkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger()
		// Refs:
		// - https://docs.cosmos.network/master/building-modules/upgrade.html#registering-migrations
		// - https://docs.cosmos.network/master/migrations/chain-upgrade-guide-044.html#chain-upgrade

		if types.IsMainnet(ctx.ChainID()) {
			if err := UpdateIBCClients(ctx, clientKeeper); err != nil {
				// log error instead of aborting the upgrade
				logger.Error("FAILED TO UPDATE IBC CLIENTS", "error", err.Error())
			}
		}

		// Leave modules are as-is to avoid running InitGenesis.
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// UpdateIBCClients updates the IBC client IDs for the Osmosis and Cosmos Hub IBC channels.
func UpdateIBCClients(ctx sdk.Context, k ibcclientkeeper.Keeper) error {
	proposalOsmosis := &ibcclienttypes.ClientUpdateProposal{
		Title:              "Update expired Osmosis IBC client",
		Description:        "Update the existing expired Cosmos Hub IBC client on Evmos (07-tendermint-0) in order to resume packet transfers between both chains.",
		SubjectClientId:    ExpiredOsmosisClient, // Osmosis Expired client
		SubstituteClientId: ActiveOsmosisClient,  // Osmosis Active client
	}

	proposalCosmosHub := &ibcclienttypes.ClientUpdateProposal{
		Title:              "Update expired Cosmos Hub IBC client",
		Description:        "Update the existing expired Cosmos Hub IBC client on Evmos (07-tendermint-3) in order to resume packet transfers between both chains.",
		SubjectClientId:    ExpiredCosmosHubClient, // Cosmos Hub Expired client
		SubstituteClientId: ActiveCosmosHubClient,  // Cosmos Hub Active client
	}

	if err := k.ClientUpdateProposal(ctx, proposalOsmosis); err != nil {
		return sdkerrors.Wrap(err, "failed to update Osmosis IBC client")
	}

	if err := k.ClientUpdateProposal(ctx, proposalCosmosHub); err != nil {
		return sdkerrors.Wrap(err, "failed to update Cosmos Hub IBC client")
	}

	return nil
}
