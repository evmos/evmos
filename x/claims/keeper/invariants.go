// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v10/x/claims/types"
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
			initialClaimablePerAction := sdk.NewDecFromInt(cr.InitialClaimableAmount).Quo(numActions)
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

		isInvariantBroken := !expectedUnclaimed.Equal(sdk.NewDecFromInt(balance.Amount))
		msg = sdk.FormatInvariant(
			types.ModuleName,
			"claims",
			fmt.Sprintf(
				"\tsum of unclaimed amount: %s\n"+
					"\tescrowed balance amount: %s\n",
				expectedUnclaimed, sdk.NewDecFromInt(balance.Amount),
			),
		)

		return msg, isInvariantBroken
	}
}
