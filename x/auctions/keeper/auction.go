// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v20/x/auctions/types"
)

// GetRound gets the current auction round
func (k *Keeper) GetRound(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixAuctionRound)
	if bz == nil {
		return 0
	}
	round := sdk.BigEndianToUint64(bz)
	return round
}

// SetRound sets the current auction round
func (k *Keeper) SetRound(ctx sdk.Context, round uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyPrefixAuctionRound, sdk.Uint64ToBigEndian(round))
}
