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
	"github.com/evmos/evmos/v10/x/inflation/types"
)

var _ types.QueryServer = Keeper{}

// Period returns the current period of the inflation module.
func (k Keeper) Period(
	c context.Context,
	_ *types.QueryPeriodRequest,
) (*types.QueryPeriodResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	period := k.GetPeriod(ctx)
	return &types.QueryPeriodResponse{Period: period}, nil
}

// EpochMintProvision returns the EpochMintProvision of the inflation module.
func (k Keeper) EpochMintProvision(
	c context.Context,
	_ *types.QueryEpochMintProvisionRequest,
) (*types.QueryEpochMintProvisionResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	epochMintProvision := k.GetEpochMintProvision(ctx)

	mintDenom := k.GetParams(ctx).MintDenom
	coin := sdk.NewDecCoinFromDec(mintDenom, epochMintProvision)

	return &types.QueryEpochMintProvisionResponse{EpochMintProvision: coin}, nil
}

// SkippedEpochs returns the number of skipped Epochs of the inflation module.
func (k Keeper) SkippedEpochs(
	c context.Context,
	_ *types.QuerySkippedEpochsRequest,
) (*types.QuerySkippedEpochsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	skippedEpochs := k.GetSkippedEpochs(ctx)
	return &types.QuerySkippedEpochsResponse{SkippedEpochs: skippedEpochs}, nil
}

// InflationRate returns the number of skipped Epochs of the inflation module.
func (k Keeper) InflationRate(
	c context.Context,
	_ *types.QueryInflationRateRequest,
) (*types.QueryInflationRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	inflationRate := k.GetInflationRate(ctx)

	return &types.QueryInflationRateResponse{InflationRate: inflationRate}, nil
}

// CirculatingSupply returns the total supply in circulation excluding the team
// allocation in the first year
func (k Keeper) CirculatingSupply(
	c context.Context,
	_ *types.QueryCirculatingSupplyRequest,
) (*types.QueryCirculatingSupplyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	circulatingSupply := k.GetCirculatingSupply(ctx)

	mintDenom := k.GetParams(ctx).MintDenom
	coin := sdk.NewDecCoinFromDec(mintDenom, circulatingSupply)

	return &types.QueryCirculatingSupplyResponse{CirculatingSupply: coin}, nil
}

// Params returns params of the mint module.
func (k Keeper) Params(
	c context.Context,
	_ *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}
