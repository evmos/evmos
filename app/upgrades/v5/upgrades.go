package v5

import (
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ibctransferkeeper "github.com/cosmos/ibc-go/v5/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"

	feemarketv010types "github.com/evmos/ethermint/x/feemarket/migrations/v010/types"
	feemarketv011 "github.com/evmos/ethermint/x/feemarket/migrations/v011"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	"github.com/evmos/evmos/v9/types"
	claimskeeper "github.com/evmos/evmos/v9/x/claims/keeper"
	claimstypes "github.com/evmos/evmos/v9/x/claims/types"
)

// TestnetDenomMetadata defines the metadata for the tEVMOS denom on testnet
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
	ck *claimskeeper.Keeper,
	sk stakingkeeper.Keeper,
	pk paramskeeper.Keeper,
	tk ibctransferkeeper.Keeper,
	xk slashingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// modify fee market parameter defaults through global
		feemarkettypes.DefaultMinGasPrice = MainnetMinGasPrices
		feemarkettypes.DefaultMinGasMultiplier = MainnetMinGasMultiplier

		// Refs:
		// - https://docs.cosmos.network/master/building-modules/upgrade.html#registering-migrations
		// - https://docs.cosmos.network/master/migrations/chain-upgrade-guide-044.html#chain-upgrade

		// define the denom metadata for the testnet
		if types.IsTestnet(ctx.ChainID()) {
			logger.Debug("setting testnet client denomination metadata...")
			bk.SetDenomMetaData(ctx, TestnetDenomMetadata)
		}

		logger.Debug("updating Tendermint consensus params...")
		UpdateConsensusParams(ctx, sk, pk)

		logger.Debug("updating IBC transfer denom traces...")
		UpdateIBCDenomTraces(ctx, tk)

		logger.Debug("swaping claims record actions...")
		ResolveAirdrop(ctx, ck)

		logger.Debug("migrating early contributor claim record...")
		MigrateContributorClaim(ctx, ck)
		// define from versions of the modules that have a new consensus version

		// migrate fee market module, other modules are left as-is to
		// avoid running InitGenesis.
		vm[feemarkettypes.ModuleName] = 2

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running migration for fee market module (EIP-1559)...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// MigrateGenesis migrates exported state from v4 to v5 genesis state.
// It performs a no-op if the migration errors.
func MigrateGenesis(appState genutiltypes.AppMap, clientCtx client.Context) genutiltypes.AppMap {
	// Migrate x/feemarket.
	if appState[feemarkettypes.ModuleName] == nil {
		return appState
	}

	// unmarshal relative source genesis application state
	var oldFeeMarketState feemarketv010types.GenesisState
	if err := clientCtx.Codec.UnmarshalJSON(appState[feemarkettypes.ModuleName], &oldFeeMarketState); err != nil {
		return appState
	}

	// delete deprecated x/feemarket genesis state
	delete(appState, feemarkettypes.ModuleName)

	// Migrate relative source genesis application state and marshal it into
	// the respective key.
	newFeeMarketState := feemarketv011.MigrateJSON(oldFeeMarketState)

	feeMarketBz, err := clientCtx.Codec.MarshalJSON(&newFeeMarketState)
	if err != nil {
		return appState
	}

	appState[feemarkettypes.ModuleName] = feeMarketBz

	return appState
}

// ResolveAirdrop iterates over all the available claim records and
// attempts to swap claimed actions and unclaimed actions.
// The following priority is considered, and only one swap will be performed:
// 1 - Unclaimed EVM      <-> claimed IBC
// 2 - Unclaimed EVM      <-> claimed Vote
// 3 - Unclaimed EVM      <-> claimed Delegate
// 4 - Unclaimed Vote     <-> claimed IBC
// 5 - Unclaimed Delegate <-> claimed IBC
//
// Rationale: A few users tokens are still locked due to issues with the airdrop
// By swapping claimed actions we allow the users to migrate the records via IBC if needed
// or mark the Evm action as completed for the Keplr users who are not able to complete it
// Since no actual claiming of action is occurring, balance will remain unchanged
func ResolveAirdrop(ctx sdk.Context, k *claimskeeper.Keeper) {
	claimsRecords := []claimstypes.ClaimsRecordAddress{}
	k.IterateClaimsRecords(ctx, func(addr sdk.AccAddress, cr claimstypes.ClaimsRecord) (stop bool) {
		// Perform any possible swap
		override := swapUnclaimedAction(cr, claimstypes.ActionEVM, claimstypes.ActionIBCTransfer) ||
			swapUnclaimedAction(cr, claimstypes.ActionEVM, claimstypes.ActionVote) ||
			swapUnclaimedAction(cr, claimstypes.ActionEVM, claimstypes.ActionDelegate) ||
			swapUnclaimedAction(cr, claimstypes.ActionVote, claimstypes.ActionIBCTransfer) ||
			swapUnclaimedAction(cr, claimstypes.ActionDelegate, claimstypes.ActionIBCTransfer)

		// If any actions were swapped, override the previous claims record
		if override {
			claim := claimstypes.ClaimsRecordAddress{
				Address:                addr.String(),
				InitialClaimableAmount: cr.InitialClaimableAmount,
				ActionsCompleted:       cr.ActionsCompleted,
			}
			claimsRecords = append(claimsRecords, claim)
		}
		return false
	})

	for _, claim := range claimsRecords {
		addr := sdk.MustAccAddressFromBech32(claim.Address)
		cr := claimstypes.ClaimsRecord{
			InitialClaimableAmount: claim.InitialClaimableAmount,
			ActionsCompleted:       claim.ActionsCompleted,
		}

		k.SetClaimsRecord(ctx, addr, cr)
	}
}

func swapUnclaimedAction(cr claimstypes.ClaimsRecord, unclaimed, claimed claimstypes.Action) bool {
	if !cr.HasClaimedAction(unclaimed) && cr.HasClaimedAction(claimed) {
		cr.ActionsCompleted[unclaimed-1] = true
		cr.ActionsCompleted[claimed-1] = false
		return true
	}
	return false
}

// MigrateContributorClaim migrates the claims record of a specific early
// contributor (Blockchain at Berkeley) from one address to another.
// See https://medium.com/evmos/the-evmos-rektdrop-abbe931ba823 for details about
// Early Contributors.
func MigrateContributorClaim(ctx sdk.Context, k *claimskeeper.Keeper) {
	from := sdk.MustAccAddressFromBech32(ContributorAddrFrom)
	to := sdk.MustAccAddressFromBech32(ContributorAddrTo)

	cr, found := k.GetClaimsRecord(ctx, from)
	if !found {
		return
	}

	k.DeleteClaimsRecord(ctx, from)
	k.SetClaimsRecord(ctx, to, cr)
}

// UpdateConsensusParams updates the Tendermint Consensus Evidence params (MaxAgeDuration and
// MaxAgeNumBlocks) to match the unbonding period and use the expected avg block time based on the
// node configuration.
func UpdateConsensusParams(ctx sdk.Context, sk stakingkeeper.Keeper, pk paramskeeper.Keeper) {
	subspace, found := pk.GetSubspace(baseapp.Paramspace)
	if !found {
		return
	}

	var evidenceParams tmproto.EvidenceParams
	subspace.GetIfExists(ctx, baseapp.ParamStoreKeyEvidenceParams, &evidenceParams)

	// safety check: no-op if the evidence params is empty (shouldn't happen)
	if evidenceParams.Equal(tmproto.EvidenceParams{}) {
		return
	}

	stakingParams := sk.GetParams(ctx)
	evidenceParams.MaxAgeDuration = stakingParams.UnbondingTime

	maxAgeNumBlocks := sdk.NewInt(int64(evidenceParams.MaxAgeDuration)).QuoRaw(int64(AvgBlockTime))
	evidenceParams.MaxAgeNumBlocks = maxAgeNumBlocks.Int64()
	subspace.Set(ctx, baseapp.ParamStoreKeyEvidenceParams, evidenceParams)
}

// UpdateIBCDenomTraces iterates over current traces to check if any of them are incorrectly formed
// and corrects the trace information.
// See https://github.com/cosmos/ibc-go/blob/main/docs/migrations/support-denoms-with-slashes.md for context.
func UpdateIBCDenomTraces(ctx sdk.Context, transferKeeper ibctransferkeeper.Keeper) {
	// list of traces that must replace the old traces in store
	var newTraces []ibctransfertypes.DenomTrace
	transferKeeper.IterateDenomTraces(ctx, func(dt ibctransfertypes.DenomTrace) bool {
		// check if the new way of splitting FullDenom
		// into Trace and BaseDenom passes validation and
		// is the same as the current DenomTrace.
		// If it isn't then store the new DenomTrace in the list of new traces.
		newTrace := ibctransfertypes.ParseDenomTrace(dt.GetFullDenomPath())
		if err := newTrace.Validate(); err == nil && !equalTraces(newTrace, dt) {
			newTraces = append(newTraces, newTrace)
		}

		return false
	})

	// replace the outdated traces with the new trace information
	for _, nt := range newTraces {
		transferKeeper.SetDenomTrace(ctx, nt)
	}
}

func equalTraces(dtA, dtB ibctransfertypes.DenomTrace) bool {
	return dtA.BaseDenom == dtB.BaseDenom && dtA.Path == dtB.Path
}
