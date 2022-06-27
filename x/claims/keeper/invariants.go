package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v6/x/claims/types"
)

// RegisterInvariants registers the claims module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "claims-invariant", k.ClaimsInvariant())
}

// ClaimsInvariant checks that the total amount of all unclaimed coins held in
// claims records is equal to the escrowed balance held in the claims module account
func (k Keeper) ClaimsInvariant() sdk.Invariant {
	return func(ctx sdk.Context) (msg string, broken bool) {
		expectedUnclaimed := sdk.ZeroDec()
		numActions := sdk.NewDec(4)
		params := k.GetParams(ctx)

		if !params.IsClaimsActive(ctx.BlockTime()) {
			return "", false
		}

		// iterate over all the claim records and sum the unclaimed amounts
		k.IterateClaimsRecords(ctx, func(_ sdk.AccAddress, cr types.ClaimsRecord) bool {
			// IMPORTANT: use Dec to prevent truncation errors
			initialClaimablePerAction := cr.InitialClaimableAmount.ToDec().Quo(numActions)
			for _, actionCompleted := range cr.ActionsCompleted {
				if !actionCompleted {
					// NOTE: only add the initial claimable amount per action for the ones that haven't been claimed
					expectedUnclaimed = expectedUnclaimed.Add(initialClaimablePerAction)
				}
			}
			return false
		})

		moduleAccAddr := k.GetModuleAccountAddress()
		balance := k.bankKeeper.GetBalance(ctx, moduleAccAddr, params.ClaimsDenom)

		isInvariantBroken := !expectedUnclaimed.Equal(balance.Amount.ToDec())
		msg = sdk.FormatInvariant(
			types.ModuleName,
			"claims",
			fmt.Sprintf(
				"\tsum of unclaimed amount: %s\n"+
					"\tescrowed balance amount: %s\n",
				expectedUnclaimed, balance.Amount.ToDec(),
			),
		)

		return msg, isInvariantBroken
	}
}
