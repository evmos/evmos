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
