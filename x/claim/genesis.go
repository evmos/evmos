package claim

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/evmos/x/claim/keeper"
	"github.com/tharsis/evmos/x/claim/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// If its the chain genesis, set the airdrop start time to be now, and setup the needed module accounts.
	if genState.Params.AirdropStartTime.Equal(time.Time{}) {
		genState.Params.AirdropStartTime = ctx.BlockTime()
		k.CreateModuleAccount(ctx, genState.ModuleAccountBalance)
	}

	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}
	if err := k.SetClaimRecords(ctx, genState.ClaimRecords); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	genesis := types.DefaultGenesis()
	genesis.ModuleAccountBalance = k.GetModuleAccountBalance(ctx)
	genesis.Params = params
	genesis.ClaimRecords = k.GetClaimRecords(ctx)
	return genesis
}
