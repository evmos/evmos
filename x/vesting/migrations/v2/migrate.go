// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	accounttypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	v1vestingtypes "github.com/evmos/evmos/v18/x/vesting/migrations/types"
	vestingtypes "github.com/evmos/evmos/v18/x/vesting/types"
)

// MigrateStore migrates the x/vesting module state from the consensus version 1 to
// version 2. Specifically, it converts all vesting accounts from their v1 proto definitions to v2.
func MigrateStore(
	ctx sdk.Context,
	ak vestingtypes.AccountKeeper,
) error {
	ak.IterateAccounts(ctx, func(account accounttypes.AccountI) bool {
		if oldAccount, ok := account.(*v1vestingtypes.ClawbackVestingAccount); ok {
			newAccount := &vestingtypes.ClawbackVestingAccount{
				BaseVestingAccount: oldAccount.BaseVestingAccount,
				FunderAddress:      oldAccount.FunderAddress,
				StartTime:          oldAccount.StartTime,
				LockupPeriods:      oldAccount.LockupPeriods,
				VestingPeriods:     oldAccount.VestingPeriods,
			}
			ak.RemoveAccount(ctx, oldAccount)
			ak.SetAccount(ctx, newAccount)
		}

		return false
	})

	return nil
}
