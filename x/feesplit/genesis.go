package feesplit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v6/x/feesplit/keeper"
	"github.com/evmos/evmos/v6/x/feesplit/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	data types.GenesisState,
) {
	k.SetParams(ctx, data.Params)

	for _, feeSplit := range data.FeeSplits {
		contract := feeSplit.GetContractAddr()
		deployer := feeSplit.GetDeployerAddr()
		withdrawer := feeSplit.GetWithdrawerAddr()

		// Set initial contracts receiving transaction fees
		k.SetFeeSplit(ctx, feeSplit)
		k.SetDeployerMap(ctx, deployer, contract)

		if len(withdrawer) != 0 {
			k.SetWithdrawerMap(ctx, withdrawer, contract)
		}
	}
}

// ExportGenesis export module state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:    k.GetParams(ctx),
		FeeSplits: k.GetFeeSplits(ctx),
	}
}
