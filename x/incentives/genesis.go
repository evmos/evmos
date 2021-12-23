package incentives

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/tharsis/evmos/x/incentives/keeper"
	"github.com/tharsis/evmos/x/incentives/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	accountKeeper authkeeper.AccountKeeper,
	data types.GenesisState,
) {
	k.SetParams(ctx, data.Params)

	// ensure incentives module account is set on genesis
	if acc := accountKeeper.GetModuleAccount(ctx, types.ModuleName); acc == nil {
		panic("the incentives module account has not been set")
	}

	for _, incentive := range data.Incentives {
		k.SetIncentive(ctx, incentive)
	}

	for _, gasMeter := range data.GasMeters {
		k.SetGasMeter(ctx, gasMeter)
	}
}

// ExportGenesis export module status
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:     k.GetParams(ctx),
		Incentives: k.GetAllIncentives(ctx),
		GasMeters:  k.GetIncentivesGasMeters(ctx),
	}
}
