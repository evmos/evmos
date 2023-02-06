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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	ethermint "github.com/evmos/evmos/v11/types"

	"github.com/evmos/evmos/v11/x/erc20/types"
)

var _ types.QueryServer = Keeper{}

// TokenPairs returns all registered pairs
func (k Keeper) TokenPairs(c context.Context, req *types.QueryTokenPairsRequest) (*types.QueryTokenPairsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	var pairs []types.TokenPair
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTokenPair)

	pageRes, err := query.Paginate(store, req.Pagination, func(_, value []byte) error {
		var pair types.TokenPair
		if err := k.cdc.Unmarshal(value, &pair); err != nil {
			return err
		}
		pairs = append(pairs, pair)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryTokenPairsResponse{
		TokenPairs: pairs,
		Pagination: pageRes,
	}, nil
}

// TokenPair returns a given registered token pair
func (k Keeper) TokenPair(c context.Context, req *types.QueryTokenPairRequest) (*types.QueryTokenPairResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// check if the token is a hex address, if not, check if it is a valid SDK
	// denom
	if err := ethermint.ValidateAddress(req.Token); err != nil {
		if err := sdk.ValidateDenom(req.Token); err != nil {
			return nil, status.Errorf(
				codes.InvalidArgument,
				"invalid format for token %s, should be either hex ('0x...') cosmos denom", req.Token,
			)
		}
	}

	id := k.GetTokenPairID(ctx, req.Token)

	if len(id) == 0 {
		return nil, status.Errorf(codes.NotFound, "token pair with token '%s'", req.Token)
	}

	pair, found := k.GetTokenPair(ctx, id)
	if !found {
		return nil, status.Errorf(codes.NotFound, "token pair with token '%s'", req.Token)
	}

	return &types.QueryTokenPairResponse{TokenPair: pair}, nil
}

// Params returns the params of the erc20 module
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}
