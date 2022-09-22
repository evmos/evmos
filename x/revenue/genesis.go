package revenue

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v8/x/revenue/keeper"
	"github.com/evmos/evmos/v8/x/revenue/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	data types.GenesisState,
) {
	k.SetParams(ctx, data.Params)

	for _, feeSplit := range data.Revenues {
		contract := feeSplit.GetContractAddr()
		deployer := feeSplit.GetDeployerAddr()
		withdrawer := feeSplit.GetWithdrawerAddr()

		// Set initial contracts receiving transaction fees
		k.SetRevenue(ctx, feeSplit)
		k.SetDeployerMap(ctx, deployer, contract)

		if len(withdrawer) != 0 {
			k.SetWithdrawerMap(ctx, withdrawer, contract)
		}
	}
}

// ExportGenesis export module state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:   k.GetParams(ctx),
		Revenues: k.GetRevenues(ctx),
	}
}
