package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tharsis/evmos/x/vesting/types"
)

var _ types.QueryServer = Keeper{}

// Unvested returns the unvested amount of tokens for a vesting account
func (k Keeper) Unvested(
	goCtx context.Context,
	req *types.QueryUnvestedRequest,
) (*types.QueryUnvestedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get vesting account
	acc := k.accountKeeper.GetAccount(ctx, addr)
	if acc == nil {
		return nil, status.Errorf(codes.NotFound,
			"account for address '%s'", req.Address,
		)
	}

	// Check if clawback vesting account
	clawbackAccount, isClawback := acc.(*types.ClawbackVestingAccount)
	if !isClawback {
		// TODO better error
		return nil, status.Errorf(codes.NotFound,
			"account for address '%s'", req.Address,
		)
	}

	unvested := clawbackAccount.GetVestingCoins(ctx.BlockTime())

	return &types.QueryUnvestedResponse{
		Unvested: unvested.String(),
	}, nil
}
