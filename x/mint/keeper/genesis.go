package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ArableProtocol/acrechain/x/mint/types"
)

// InitGenesis new mint genesis.
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	if data == nil {
		panic("empty mint genesis state")
	}

	minter := types.Minter{
		DailyProvisions: data.Params.GenesisDailyProvisions,
		LastMintTime:    ctx.BlockTime().Unix(),
	}
	k.SetMinter(ctx, minter)
	k.SetParams(ctx, data.Params)

	// The call to GetModuleAccount creates a module account if it does not exist.
	k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	k.SetNextReductionTime(ctx, data.Params.NextRewardsReductionTime)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)

	return types.NewGenesisState(params)
}
