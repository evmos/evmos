// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v18/x/erc20/types"
)

// SetSTRv2Address stores an address that will be affected by the
// Single Token Representation v2 migration.
func (k Keeper) SetSTRv2Address(ctx sdk.Context, address sdk.AccAddress) {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.KeyPrefixSTRv2Addresses,
	)
	store.Set(address.Bytes(), []byte{})
}

// HasSTRv2Address checks if a given address has already been stored as
// affected by the STR v2 migration.
func (k Keeper) HasSTRv2Address(ctx sdk.Context, address sdk.AccAddress) bool {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.KeyPrefixSTRv2Addresses,
	)
	return store.Has(address.Bytes())
}

// DeleteSTRv2Address removes the entry already stored
// NOTE: for testing purpose only
func (k Keeper) DeleteSTRv2Address(ctx sdk.Context, address sdk.AccAddress) {
	store := prefix.NewStore(
		ctx.KVStore(k.storeKey),
		types.KeyPrefixSTRv2Addresses,
	)
	store.Delete(address.Bytes())
}

// GetAllSTRV2Address iterates over all the stored accounts that interacted with registered coins.
// and returns them in an array
func (k Keeper) GetAllSTRV2Address(ctx sdk.Context) []sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixSTRv2Addresses)
	defer iterator.Close()

	accAddresses := []sdk.AccAddress{}

	for ; iterator.Valid(); iterator.Next() {
		// First byte is the prefix, final bytes is the address
		// iterator.Value is empty
		address := sdk.AccAddress(iterator.Key()[1:])
		accAddresses = append(accAddresses, address)
	}
	return accAddresses
}
