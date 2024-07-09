// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v18/x/auctions/types"
)

// GetEpochIdentifier gets the epoch identifier
func (k Keeper) GetEpochIdentifier(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixEpochIdentifier)
	if len(bz) == 0 {
		return ""
	}

	return string(bz)
}

// SetEpochIdentifier stores the epoch identifier
func (k Keeper) SetEpochIdentifier(ctx sdk.Context, epochIdentifier string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyPrefixEpochIdentifier, []byte(epochIdentifier))
}
