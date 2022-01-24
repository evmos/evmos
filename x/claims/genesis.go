package claims

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/evmos/x/claims/keeper"
	"github.com/tharsis/evmos/x/claims/types"
)

// InitGenesis initializes the claim module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data types.GenesisState) {
	totalEscrowed := sdk.ZeroInt()
	sumClaims := sdk.ZeroInt()

	// ensure claim module account is set on genesis
	if acc := k.GetModuleAccountAccount(ctx); acc == nil {
		panic("the claim module account has not been set")
	}

	// set the start time to the current block time by default
	if data.Params.AirdropStartTime.IsZero() {
		data.Params.AirdropStartTime = ctx.BlockTime()
	}

	k.SetParams(ctx, data.Params)

	escrowedCoins := k.GetModuleAccountBalances(ctx)
	if escrowedCoins != nil {
		totalEscrowed = escrowedCoins.AmountOfNoDenomValidation(data.Params.ClaimDenom)
	}

	for _, claimRecord := range data.ClaimRecords {
		addr, _ := sdk.AccAddressFromBech32(claimRecord.Address)
		cr := types.ClaimRecord{
			InitialClaimableAmount: claimRecord.InitialClaimableAmount,
			ActionsCompleted:       claimRecord.ActionsCompleted,
		}
		k.SetClaimRecord(ctx, addr, cr)

		sumClaims = sumClaims.Add(cr.InitialClaimableAmount)
	}

	if !sumClaims.Equal(totalEscrowed) {
		panic(
			fmt.Errorf(
				"sum of claimable amount ≠ escrowed module account amount (%s ≠ %s)",
				sumClaims, totalEscrowed,
			),
		)
	}
}

// ExportGenesis returns the claim module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:       k.GetParams(ctx),
		ClaimRecords: k.GetClaimRecords(ctx),
	}
}
