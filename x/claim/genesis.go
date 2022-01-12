package claim

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/evmos/x/claim/keeper"
	"github.com/tharsis/evmos/x/claim/types"
)

// InitGenesis initializes the claim module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data types.GenesisState) {
	k.SetParams(ctx, data.Params)

	escrowedCoin := sdk.Coin{Denom: data.Params.ClaimDenom, Amount: sdk.ZeroInt()}
	escrowedCoins := k.GetModuleAccountBalances(ctx)
	if escrowedCoins != nil {
		escrowedCoin.Amount = escrowedCoins.AmountOfNoDenomValidation(data.Params.ClaimDenom)
	}

	totalToClaim := sdk.Coin{Denom: data.Params.ClaimDenom, Amount: sdk.ZeroInt()}

	for _, claimRecord := range data.ClaimRecords {
		addr, _ := sdk.AccAddressFromBech32(claimRecord.Address)
		cr := types.ClaimRecord{
			InitialClaimableAmount: claimRecord.InitialClaimableAmount,
			ActionsCompleted:       claimRecord.ActionsCompleted,
		}
		k.SetClaimRecord(ctx, addr, cr)

		amt := sdk.Coin{Denom: data.Params.ClaimDenom, Amount: cr.InitialClaimableAmount}
		totalToClaim = totalToClaim.Add(amt)
	}

	if !totalToClaim.IsEqual(escrowedCoin) {
		panic(
			fmt.Errorf(
				"sum of claimable amount ≠ escrowed module account amount (%s ≠ %s)",
				totalToClaim, escrowedCoin,
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
