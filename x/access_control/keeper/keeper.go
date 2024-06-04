// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) Keeper {
	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
	}
}
