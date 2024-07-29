package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/x/auctions/types"
)

// getRound gets the current auction round
func (k *Keeper) getRound(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixAuctionRound)
	if bz == nil {
		return 0
	}
	round := sdk.BigEndianToUint64(bz)
	return round
}

// setRound sets the current auction round
func (k *Keeper) setRound(ctx sdk.Context, round uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyPrefixAuctionRound, sdk.Uint64ToBigEndian(round))
}
