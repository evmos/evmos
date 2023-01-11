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
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/evmos/evmos/v11/x/vesting/types"
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

	// Get vesting account
	acc := k.accountKeeper.GetAccount(ctx, addr)
	if acc == nil {
		return nil, status.Errorf(
			codes.NotFound,
			"account for address '%s'", req.Address,
		)
	}

	// Check if clawback vesting account
	clawbackAccount, isClawback := acc.(*types.ClawbackVestingAccount)
	if !isClawback {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"account at address '%s' is not a vesting account ", req.Address,
		)
	}

	locked := clawbackAccount.GetLockedOnly(ctx.BlockTime())
	unvested := clawbackAccount.GetUnvestedOnly(ctx.BlockTime())
	vested := clawbackAccount.GetVestedOnly(ctx.BlockTime())

	return &types.QueryBalancesResponse{
		Locked:   locked,
		Unvested: unvested,
		Vested:   vested,
	}, nil
}
