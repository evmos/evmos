// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	accounttypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/evmos/evmos/v13/utils"
	v1vestingtypes "github.com/evmos/evmos/v13/x/vesting/migrations/types"
	vestingtypes "github.com/evmos/evmos/v13/x/vesting/types"
	"github.com/tendermint/tendermint/libs/log"
)

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
	logger := k.Logger(ctx)

	ak.IterateAccounts(ctx, func(account accounttypes.AccountI) bool {
		if oldAccount, ok := account.(*v1vestingtypes.ClawbackVestingAccount); ok {
			if utils.IsMainnet(ctx.ChainID()) {
				newAccount := &vestingtypes.ClawbackVestingAccount{
					BaseVestingAccount: oldAccount.BaseVestingAccount,
					FunderAddress:      oldAccount.FunderAddress,
					StartTime:          oldAccount.StartTime,
					LockupPeriods:      oldAccount.LockupPeriods,
					VestingPeriods:     oldAccount.VestingPeriods,
				}
				ak.RemoveAccount(ctx, oldAccount)
				ak.SetAccount(ctx, newAccount)
				k.SetGovClawbackEnabled(ctx, newAccount.GetAddress())
				logger.Debug("enabled clawback via governance", "address", newAccount.Address)
			}
			k.SetGovClawbackEnabled(ctx, oldAccount.GetAddress())
			logger.Debug("enabled clawback via governance", "address", oldAccount.Address)

		}

		return false
	})

	return nil
}
