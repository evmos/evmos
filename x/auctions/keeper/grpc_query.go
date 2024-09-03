// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/x/auctions/types"
)

var _ types.QueryServer = Keeper{}

// AuctionInfo returns the current auction information
func (k Keeper) AuctionInfo(c context.Context, _ *types.QueryCurrentAuctionInfoRequest) (*types.QueryCurrentAuctionInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if !k.GetParams(ctx).EnableAuction {
		return nil, types.ErrAuctionDisabled
	}

	moduleAddress := k.accountKeeper.GetModuleAddress(types.ModuleName)
	coins := k.bankKeeper.GetAllBalances(ctx, moduleAddress)

	// Filter out the coin with the specified denomination
	filteredCoins := removeBaseCoinFromCoins(coins)

	currentRound := k.GetRound(ctx)
	highestBid := k.GetHighestBid(ctx)
	return &types.QueryCurrentAuctionInfoResponse{
		Tokens:        filteredCoins,
		CurrentRound:  currentRound,
		HighestBid:    highestBid.BidValue,
		BidderAddress: highestBid.Sender,
	}, nil
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
