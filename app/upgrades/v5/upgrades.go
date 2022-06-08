package v5

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	feemarketv010types "github.com/tharsis/ethermint/x/feemarket/migrations/v010/types"
	feemarketv011 "github.com/tharsis/ethermint/x/feemarket/migrations/v011"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	"github.com/tharsis/evmos/v5/types"
	claimskeeper "github.com/tharsis/evmos/v5/x/claims/keeper"
	claimstypes "github.com/tharsis/evmos/v5/x/claims/types"
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
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Refs:
		// - https://docs.cosmos.network/master/building-modules/upgrade.html#registering-migrations
		// - https://docs.cosmos.network/master/migrations/chain-upgrade-guide-044.html#chain-upgrade

		// define the denom metadata for the testnet
		if types.IsTestnet(ctx.ChainID()) {
			bk.SetDenomMetaData(ctx, TestnetDenomMetadata)
		}

		if types.IsMainnet(ctx.ChainID()) {
			ResolveAirdrop(ctx, ck)
		}

		// define from versions of the modules that have a new consensus version

		// migrate fee market module, other modules are left as-is to
		// avoid running InitGenesis.
		vm[feemarkettypes.ModuleName] = 3

		// Leave modules are as-is to avoid running InitGenesis.
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

// ResolveAirdrop will iterate over all the available claim records and
// attempt to swap claimed actions and unclaimed actions.
// The following priority is considered, and only one swap will be performed:
// 1 - Unclaimed EVM      <-> claimed IBC
// 2 - Unclaimed EVM      <-> claimed Vote
// 3 - Unclaimed EVM      <-> claimed Delegate
// 4 - Unclaimed Vote     <-> claimed IBC
// 5 - Unclaimed Delegate <-> claimed IBC
func ResolveAirdrop(ctx sdk.Context, k *claimskeeper.Keeper) {
	claimsRecords := []claimstypes.ClaimsRecordAddress{}
	k.IterateClaimsRecords(ctx, func(addr sdk.AccAddress, cr claimstypes.ClaimsRecord) (stop bool) {
		//Perform any possible swap
		override := swapActionsIfUncompleted(cr, claimstypes.ActionEVM, claimstypes.ActionIBCTransfer) ||
			swapActionsIfUncompleted(cr, claimstypes.ActionEVM, claimstypes.ActionVote) ||
			swapActionsIfUncompleted(cr, claimstypes.ActionEVM, claimstypes.ActionDelegate) ||
			swapActionsIfUncompleted(cr, claimstypes.ActionVote, claimstypes.ActionIBCTransfer) ||
			swapActionsIfUncompleted(cr, claimstypes.ActionDelegate, claimstypes.ActionIBCTransfer)

		// If any actions were swapped, override the previous claims record
		if override {
			cra := claimstypes.ClaimsRecordAddress{
				Address:                addr.String(),
				InitialClaimableAmount: cr.InitialClaimableAmount,
				ActionsCompleted:       cr.ActionsCompleted,
			}
			claimsRecords = append(claimsRecords, cra)
		}
		return false
	})

	for _, claim := range claimsRecords {
		ad, _ := sdk.AccAddressFromBech32(claim.Address)
		k.SetClaimsRecord(ctx, ad,
			claimstypes.ClaimsRecord{InitialClaimableAmount: claim.InitialClaimableAmount,
				ActionsCompleted: claim.ActionsCompleted})
	}
}

func swapActionsIfUncompleted(cr claimstypes.ClaimsRecord, uncompleted, completed claimstypes.Action) bool {
	if !cr.HasClaimedAction(uncompleted) && cr.HasClaimedAction(completed) {
		cr.ActionsCompleted[uncompleted-1] = true
		cr.ActionsCompleted[completed-1] = false
		return true
	}
	return false
}
