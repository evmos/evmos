package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v18/utils"
	"github.com/evmos/evmos/v18/x/auctions/types"
)

// GetHighestBid gets the highest bid
func (k *Keeper) GetHighestBid(ctx sdk.Context) *types.Bid {
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

// SetHighestBid sets the highest bid
func (k *Keeper) SetHighestBid(ctx sdk.Context, sender string, amount sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	bid := &types.Bid{
		Sender: sender,
		Amount: amount,
	}
	bz := k.cdc.MustMarshal(bid)
	store.Set(types.KeyPrefixBid, bz)
}

// DeleteBid deletes the highest bid
func (k *Keeper) DeleteBid(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyPrefixBid)
}
