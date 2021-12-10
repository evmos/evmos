package distribution

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/tharsis/evmos/x/distribution/keeper"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	accountKeeper authkeeper.AccountKeeper,
	// data types.GenesisState,
) {
	// k.SetParams(ctx, data.Params)
}

// ExportGenesis export module status
// func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
// 	return &types.GenesisState{
// 		Params: k.GetParams(ctx),
// 	}
// }
