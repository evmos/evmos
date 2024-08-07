// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingtypes "github.com/evmos/evmos/v19/x/vesting/types"
)

// MigrateStore migrates the x/vesting module state from the consensus version 2 to
// version 3.
// Specifically, it adds the DelegatedVesting (that should always be 0)
// to the DelegatedFree
func MigrateStore(
	ctx sdk.Context,
	ak vestingtypes.AccountKeeper,
) error {
	ak.IterateAccounts(ctx, func(account sdk.AccountI) bool {
		if vestAcc, ok := account.(*vestingtypes.ClawbackVestingAccount); ok {
			// if DelegatedVesting == 0, skip it
			if !vestAcc.DelegatedVesting.IsAllPositive() {
				return false
			}
			// add DelegatedVesting to DelegatedFree,
			// because it is not possible to delegate vesting coins.
			// ONLY vested (free) coins can be delegated
			vestAcc.DelegatedFree = vestAcc.DelegatedFree.Add(vestAcc.DelegatedVesting...)
			vestAcc.DelegatedVesting = sdk.NewCoins()

			ak.SetAccount(ctx, vestAcc)
		}

		return false
	})

	return nil
}
