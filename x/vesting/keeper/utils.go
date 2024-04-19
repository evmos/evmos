// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/evmos/evmos/v17/x/vesting/types"
)

// GetClawbackVestingAccount is a helper function to get the account from the
// account keeper and check if it is of the correct type for clawback vesting.
func (k Keeper) GetClawbackVestingAccount(ctx sdk.Context, addr sdk.AccAddress) (*types.ClawbackVestingAccount, error) {
	acc := k.accountKeeper.GetAccount(ctx, addr)
	if acc == nil {
		return nil, errorsmod.Wrapf(errortypes.ErrUnknownAddress, "account at address '%s' does not exist", addr.String())
	}

	clawbackAccount, isClawback := acc.(*types.ClawbackVestingAccount)
	if !isClawback {
		return nil, errorsmod.Wrap(types.ErrNotSubjectToClawback, addr.String())
	}

	return clawbackAccount, nil
}
