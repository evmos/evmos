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

	// Ensure unvested team module account is set on genesis
	if acc := ak.GetModuleAccount(ctx, types.UnvestedTeamAccount); acc == nil {
		panic("the unvested team module account has not been set")
	}

	// TODO Ensure team account is set on genesis
	// acc := ak.GetAccount(ctx, sdk.AccAddress(data.Params.TeamAddress))
	// if acc.GetAddress().Empty() {
	// 	panic("the team account has not been set")
	// }

	// Set Period
	period := data.Period
	k.SetPeriod(ctx, period)

	// Calculate epoch mint provision
	epochMintProvision := types.CalculateEpochMintProvision(data.Params, period)
	k.SetEpochMintProvision(ctx, epochMintProvision)

	// Mint genesis coins for teamVesting
	amount := sdk.NewInt(200_000_000)
	coins := sdk.NewCoins(sdk.NewCoin(data.Params.MintDenom, amount))
	if err := k.MintGenesisTeamVestingCoins(ctx, coins); err != nil {
		panic(err)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
		Period: k.GetPeriod(ctx),
	}
}
