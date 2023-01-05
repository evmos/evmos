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

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/evmos/evmos/v10/x/epochs/types"
)

var _ types.QueryServer = Keeper{}

// EpochInfos provide running epochInfos
func (k Keeper) EpochInfos(
	c context.Context,
	req *types.QueryEpochsInfoRequest,
) (*types.QueryEpochsInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	var epochs []types.EpochInfo
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixEpoch)

	pageRes, err := query.Paginate(store, req.Pagination, func(_, value []byte) error {
		var epoch types.EpochInfo
		if err := k.cdc.Unmarshal(value, &epoch); err != nil {
			return err
		}
		epochs = append(epochs, epoch)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryEpochsInfoResponse{
		Epochs:     epochs,
		Pagination: pageRes,
	}, nil
}

// CurrentEpoch provides current epoch of specified identifier
func (k Keeper) CurrentEpoch(
	c context.Context,
	req *types.QueryCurrentEpochRequest,
) (*types.QueryCurrentEpochResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	info, found := k.GetEpochInfo(ctx, req.Identifier)
	if !found {
		return nil, status.Errorf(codes.NotFound, "epoch info not found: %s", req.Identifier)
	}

	return &types.QueryCurrentEpochResponse{
		CurrentEpoch: info.CurrentEpoch,
	}, nil
}
