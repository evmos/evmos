package intrarelayer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/intrarelayer/keeper"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

// InitGenesis import module genesis
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data types.GenesisState) {
	k.SetParams(ctx, data.Params)

	for _, pair := range data.TokenPairs {
		k.SetTokenPair(ctx, pair)
	}
}

// ExportGenesis export module status
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:     k.GetParams(ctx),
		TokenPairs: k.GetAllTokenPairs(ctx),
	}
}
