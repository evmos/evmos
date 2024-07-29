package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v18/utils"
	"github.com/evmos/evmos/v18/x/auctions/types"
)

var _ types.QueryServer = Keeper{}

// AuctionInfo returns the current auction information
func (k Keeper) AuctionInfo(c context.Context, _ *types.QueryCurrentAuctionInfoRequest) (*types.QueryCurrentAuctionInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	moduleAddress := k.accountKeeper.GetModuleAddress(types.ModuleName)
	coins := k.bankKeeper.GetAllBalances(ctx, moduleAddress)

	// Filter out the coin with the specified denomination
	filteredCoins := sdk.Coins{}
	for _, coin := range coins {
		if coin.Denom != utils.BaseDenom {
			filteredCoins = append(filteredCoins, coin)
		}
	}

	currentRound := k.getRound(ctx)
	highestBid := k.getHighestBid(ctx)
	return &types.QueryCurrentAuctionInfoResponse{
		Tokens:        filteredCoins,
		CurrentRound:  currentRound,
		HighestBid:    highestBid.Amount,
		BidderAddress: highestBid.Sender,
	}, nil
}

// Params returns params of the auctions module.
func (k Keeper) Params(
	c context.Context,
	_ *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.getParams(ctx)
	return &types.QueryParamsResponse{Params: &params}, nil
}
