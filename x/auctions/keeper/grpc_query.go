package keeper

import (
	"context"
	"github.com/evmos/evmos/v18/x/auctions/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) AuctionTokens(ctx context.Context, request *types.QueryAuctionTokensRequest) (*types.QueryAuctionTokensResponse, error) {
	return nil, nil
}
