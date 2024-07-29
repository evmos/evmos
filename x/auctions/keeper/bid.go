package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/auctions/types"
)

// getHighestBid gets the highest bid
func (k *Keeper) getHighestBid(ctx sdk.Context) *types.Bid {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixBid)

	if bz == nil {
		return &types.Bid{
			Sender: "",
			Amount: sdk.NewCoin(utils.BaseDenom, sdk.ZeroInt()),
		}
	}

	var bid types.Bid
	k.cdc.MustUnmarshal(bz, &bid)
	return &bid
}

// setHighestBid sets the highest bid
func (k *Keeper) setHighestBid(ctx sdk.Context, sender string, amount sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	bid := &types.Bid{
		Sender: sender,
		Amount: amount,
	}
	bz := k.cdc.MustMarshal(bid)
	store.Set(types.KeyPrefixBid, bz)
}

// deleteBid deletes the highest bid
func (k *Keeper) deleteBid(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyPrefixBid)
}
