// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/evmos/evmos/v18/x/vesting/types"
)

var _ types.QueryServer = Keeper{}

// Balances returns the locked, unvested and vested amount of tokens for a
// clawback vesting account
func (k Keeper) Balances(
	goCtx context.Context,
	req *types.QueryBalancesRequest,
) (*types.QueryBalancesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	clawbackAccount, err := k.GetClawbackVestingAccount(ctx, addr)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"account at address '%s' either does not exist or is not a vesting account ", addr.String(),
		)
	}

	locked := clawbackAccount.GetLockedUpCoins(ctx.BlockTime())
	unvested := clawbackAccount.GetVestingCoins(ctx.BlockTime())
	vested := clawbackAccount.GetVestedCoins(ctx.BlockTime())

	return &types.QueryBalancesResponse{
		Locked:   locked,
		Unvested: unvested,
		Vested:   vested,
	}, nil
}
