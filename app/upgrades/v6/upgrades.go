package v6

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ibctransferkeeper "github.com/cosmos/ibc-go/v3/modules/apps/transfer/keeper"

	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	v5 "github.com/evmos/evmos/v9/app/upgrades/v5"
	"github.com/evmos/evmos/v9/types"
	claimskeeper "github.com/evmos/evmos/v9/x/claims/keeper"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v6
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bk bankkeeper.Keeper,
	ck *claimskeeper.Keeper,
	sk stakingkeeper.Keeper,
	pk paramskeeper.Keeper,
	tk ibctransferkeeper.Keeper,
	xk slashingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// modify fee market parameter defaults through global
		feemarkettypes.DefaultMinGasPrice = v5.MainnetMinGasPrices
		feemarkettypes.DefaultMinGasMultiplier = v5.MainnetMinGasMultiplier

		// Refs:
		// - https://docs.cosmos.network/master/building-modules/upgrade.html#registering-migrations
		// - https://docs.cosmos.network/master/migrations/chain-upgrade-guide-044.html#chain-upgrade

		// Mainnet is migrating from v4 -> v6 so we need to add the v5 migration logic
		if types.IsMainnet(ctx.ChainID()) {
			logger.Debug("updating Tendermint consensus params...")
			v5.UpdateConsensusParams(ctx, sk, pk)

			logger.Debug("updating slashing period...")
			UpdateSlashingParams(ctx, xk)

			logger.Debug("updating IBC transfer denom traces...")
			v5.UpdateIBCDenomTraces(ctx, tk)

			logger.Debug("swaping claims record actions...")
			v5.ResolveAirdrop(ctx, ck)

			logger.Debug("migrating early contributor claim record...")
			v5.MigrateContributorClaim(ctx, ck)

			// define from versions of the modules that have a new consensus version
			// migrate fee market module
			vm[feemarkettypes.ModuleName] = 2
		}

		// migrate EVM module from v1 -> v2
		vm[evmtypes.ModuleName] = 1

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// UpdateSlashingParams updates the Slashing params (SignedBlocksWindow) to
// increase to keep the same wall-time of reaction time, since the block times
// are expected to be 67% shorter.
func UpdateSlashingParams(ctx sdk.Context, sk slashingkeeper.Keeper) {
	params := sk.GetParams(ctx)
	params.SignedBlocksWindow *= 3 // migrate from mainnet from 30,000 -> 90,000
	sk.SetParams(ctx, params)
}
