package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v14/x/vesting/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetClawbackVestingAccount is a helper function to get the account from the
// account keeper and check if it is of the correct type for clawback vesting.
func (k Keeper) GetClawbackVestingAccount(ctx sdk.Context, addr sdk.AccAddress) (*types.ClawbackVestingAccount, error) {
	acc := k.accountKeeper.GetAccount(ctx, addr)
	if acc == nil {
		// TODO: should we use errortypes.ErrInvalidRequest here?
		return nil, status.Errorf(
			codes.NotFound,
			"account for address '%s'", addr.String(),
		)
	}

	clawbackAccount, isClawback := acc.(*types.ClawbackVestingAccount)
	if !isClawback {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"account at address '%s' is not a vesting account ", addr.String(),
		)
	}

	return clawbackAccount, nil
}
