// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

// Keeper of the inflation store
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
}

// NewKeeper creates a new mint Keeper instance
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
) Keeper {
	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
	}
}
