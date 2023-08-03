// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	accounttypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	v1vestingtypes "github.com/evmos/evmos/v13/x/vesting/migrations/types"
	vestingtypes "github.com/evmos/evmos/v13/x/vesting/types"
	"github.com/tendermint/tendermint/libs/log"
)

var addresses = []string{
	"evmos19mqtl7pyvtazl85jlre9jltpuff9enjdn9m7hz",
}

// VestingKeeper defines the expected keeper for vesting
type VestingKeeper interface {
	Logger(ctx sdk.Context) log.Logger
	SetGovClawbackEnabled(ctx sdk.Context, address sdk.AccAddress)
}

// MigrateStore migrates the x/vesting module state from the consensus version 1 to
// version 2. Specifically, it converts all vesting accounts from their v1 proto definitions to v2.
func MigrateStore(
	ctx sdk.Context,
	k VestingKeeper,
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

	logger := k.Logger(ctx)

	for _, addr := range addresses {
		accAddres := sdk.MustAccAddressFromBech32(addr)
		k.SetGovClawbackEnabled(ctx, accAddres)
		logger.Debug("enabled clawback via governance", "address", addr)
	}

	return nil
}
