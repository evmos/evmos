package inflation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/inflation/keeper"
	"github.com/tharsis/evmos/x/inflation/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	ak types.AccountKeeper,
	data types.GenesisState,
) {
	k.SetParams(ctx, data.Params)

	// Ensure inflation module account is set on genesis
	if acc := ak.GetModuleAccount(ctx, types.ModuleName); acc == nil {
		panic("the inflation module account has not been set")
	}

	period := data.Period
	k.SetPeriod(ctx, period)

	epochMintProvisions := types.CalculateEpochMintProvisions(data.Params, period)
	k.SetEpochMintProvision(ctx, epochMintProvisions)

	// TODO mint team vesting coins
	// Mint initial coins for teamVesting
	// initialTeamVestingCoins := sdk.NewCoin(data.Params.MintDenom, sdk.NewInt(200_000_000))
	// k.MintInitialTeamVestingCoins(ctx, initialTeamVestingCoins)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
		Period: k.GetPeriod(ctx),
	}
}
