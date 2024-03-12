// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v16/x/erc20/types"
)

// SetSTRv2Address stores an address that will be affected by the
// Single Token Representation v2 migration.
func (k Keeper) SetSTRv2Address(ctx sdk.Context, address common.Address) {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.KeyPrefixSTRv2Addresses,
	)
	store.Set(address.Bytes(), []byte{})
}

// HasSTRv2Address checks if a given address has already been stored as
// affected by the STR v2 migration.
func (k Keeper) HasSTRv2Address(ctx sdk.Context, address common.Address) bool {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.KeyPrefixSTRv2Addresses,
	)
	return store.Has(address.Bytes())
}
