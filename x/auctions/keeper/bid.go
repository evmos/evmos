// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/auctions/types"
)

// GetHighestBid gets the highest bid
func (k *Keeper) GetHighestBid(ctx sdk.Context) types.Bid {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixBid)

	if bz == nil {
		return types.Bid{
			Sender:   "",
			BidValue: sdk.NewCoin(utils.BaseDenom, sdk.ZeroInt()),
		}
	}

	var bid types.Bid
	k.cdc.MustUnmarshal(bz, &bid)
	return bid
}

// SetHighestBid sets the highest bid
func (k *Keeper) SetHighestBid(ctx sdk.Context, sender string, amount sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	bid := &types.Bid{
		Sender:   sender,
		BidValue: amount,
	}
	bz := k.cdc.MustMarshal(bid)
	store.Set(types.KeyPrefixBid, bz)
}

// deleteBid deletes the highest bid
func (k *Keeper) deleteBid(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyPrefixBid)
}

// refundLastBid refunds the last bid placed on an auction
func (k Keeper) refundLastBid(ctx sdk.Context) error {
	lastBid := k.GetHighestBid(ctx)
	lastBidder, err := sdk.AccAddressFromBech32(lastBid.Sender)
	if err != nil {
		return err
	}
	bidAmount := sdk.NewCoins(lastBid.BidValue)
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lastBidder, bidAmount)
}
