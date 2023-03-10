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

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v12/x/feemarket/types"
)

var _ types.QueryServer = Keeper{}

// Params implements the Query/Params gRPC method
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// BaseFee implements the Query/BaseFee gRPC method
func (k Keeper) BaseFee(c context.Context, _ *types.QueryBaseFeeRequest) (*types.QueryBaseFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	res := &types.QueryBaseFeeResponse{}
	baseFee := k.GetBaseFee(ctx)

	if baseFee != nil {
		aux := sdkmath.NewIntFromBigInt(baseFee)
		res.BaseFee = &aux
	}

	return res, nil
}

// BlockGas implements the Query/BlockGas gRPC method
func (k Keeper) BlockGas(c context.Context, _ *types.QueryBlockGasRequest) (*types.QueryBlockGasResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	gas := sdkmath.NewIntFromUint64(k.GetBlockGasWanted(ctx))

	if !gas.IsInt64() {
		return nil, errorsmod.Wrapf(sdk.ErrIntOverflowCoin, "block gas %s is higher than MaxInt64", gas)
	}

	return &types.QueryBlockGasResponse{
		Gas: gas.Int64(),
	}, nil
}
