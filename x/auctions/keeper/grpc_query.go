package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v18/x/auctions/types"
)

var _ types.QueryServer = Keeper{}

// AuctionTokens returns the current module account assets that are being auctioned.
func (k Keeper) AuctionTokens(c context.Context, _ *types.QueryAuctionTokensRequest) (*types.QueryAuctionTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	moduleAddress := k.accountKeeper.GetModuleAddress(types.ModuleName)
	coins := k.bankKeeper.GetAllBalances(ctx, moduleAddress)

	return &types.QueryAuctionTokensResponse{Amount: coins}, nil
}

// Params returns params of the auctions module.
func (k Keeper) Params(
	c context.Context,
	_ *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: &params}, nil
}
