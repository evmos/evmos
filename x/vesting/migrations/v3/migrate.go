// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	accounttypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/evmos/evmos/v14/x/vesting/types"
	"github.com/tendermint/tendermint/libs/log"
)

// VestingKeeper defines the expected keeper for vesting
type VestingKeeper interface {
	Logger(ctx sdk.Context) log.Logger
	SetGovClawbackEnabled(ctx sdk.Context, address sdk.AccAddress)
}

// MigrateStore migrates the x/vesting module state from the consensus version 2 to
// version 3. Specifically, it enables clawback for all vesting accounts.
func MigrateStore(
	ctx sdk.Context,
	k VestingKeeper,
	ak vestingtypes.AccountKeeper,
) error {
	logger := k.Logger(ctx)

	ak.IterateAccounts(ctx, func(account accounttypes.AccountI) bool {
		if oldAccount, ok := account.(*vestingtypes.ClawbackVestingAccount); ok {
			k.SetGovClawbackEnabled(ctx, oldAccount.GetAddress())
			logger.Debug("enabled clawback via governance", "address", oldAccount.Address)
		}

		return false
	})

	return nil
}
