package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/evmos/v2/x/claims/types"
)

// RegisterInvariants registers the claims module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "claims-invariant", k.ClaimsInvariant())
}

// ClaimsInvariant checks that the escrowed balance amount reflects all the
// unclaimed held in claims records
func (k Keeper) ClaimsInvariant() sdk.Invariant {
	return func(ctx sdk.Context) (msg string, broken bool) {
		expectedUnclaimed := sdk.ZeroInt()
		numActions := sdk.NewInt(4)
		params := k.GetParams(ctx)

		if !params.IsClaimsActive(ctx.BlockTime()) {
			return "", false
		}

		//
		k.IterateClaimsRecords(ctx, func(_ sdk.AccAddress, cr types.ClaimsRecord) bool {
			initialClaimablePerAction := cr.InitialClaimableAmount.Quo(numActions)

			for _, actionCompleted := range cr.ActionsCompleted {
				if !actionCompleted {
					// NOTE: only add the initial claimable amount per action for the ones that haven't been claimed
					expectedUnclaimed = expectedUnclaimed.Add(initialClaimablePerAction)
				}
			}
			return false
		})

		moduleAccAddr := k.GetModuleAccountAddress(ctx)
		balance := k.bankKeeper.GetBalance(ctx, moduleAccAddr, params.ClaimsDenom)

		isInvariantBroken := !expectedUnclaimed.Equal(balance.Amount)
		msg = sdk.FormatInvariant(
			types.ModuleName,
			"claims",
			fmt.Sprintf(
				"\tsum of unclaimed amount: %s\n"+
					"\tescrowed balance amount: %s\n",
				expectedUnclaimed, balance.Amount,
			),
		)

		return msg, isInvariantBroken
	}
}
