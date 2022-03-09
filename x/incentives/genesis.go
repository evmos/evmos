package incentives

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/tharsis/evmos/v2/x/incentives/keeper"
	"github.com/tharsis/evmos/v2/x/incentives/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	accountKeeper authkeeper.AccountKeeper,
	data types.GenesisState,
) {
	k.SetParams(ctx, data.Params)

	// Ensure incentives module account is set on genesis
	if acc := accountKeeper.GetModuleAccount(ctx, types.ModuleName); acc == nil {
		panic("the incentives module account has not been set")
	}

	allocationMeters := make(map[string]sdk.Dec)

	for _, incentive := range data.Incentives {
		// Set Incentives
		k.SetIncentive(ctx, incentive)

		// Build allocation meter map
		for _, al := range incentive.Allocations {
			allocationMeters[al.Denom] = allocationMeters[al.Denom].Add(al.Amount)
		}
	}

	// Set allocation meters
	for denom, amount := range allocationMeters {
		am := sdk.DecCoin{
			Denom:  denom,
			Amount: amount,
		}
		k.SetAllocationMeter(ctx, am)
	}

	// Set gas meters
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
