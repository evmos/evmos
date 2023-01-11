// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v11/x/claims/types"
)

// GetClaimsRecord returns the claims record for a specific address
func (k Keeper) GetClaimsRecord(ctx sdk.Context, addr sdk.AccAddress) (types.ClaimsRecord, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixClaimsRecords)

	bz := store.Get(addr)
	if len(bz) == 0 {
		return types.ClaimsRecord{}, false
	}

	var claimsRecord types.ClaimsRecord
	k.cdc.MustUnmarshal(bz, &claimsRecord)

	return claimsRecord, true
}

// HasClaimsRecord returns if the claims record is found in the store a
// given address
func (k Keeper) HasClaimsRecord(ctx sdk.Context, addr sdk.AccAddress) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixClaimsRecords)
	return store.Has(addr)
}

// SetClaimsRecord sets a claims record for an address in store
func (k Keeper) SetClaimsRecord(ctx sdk.Context, addr sdk.AccAddress, claimsRecord types.ClaimsRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixClaimsRecords)
	bz := k.cdc.MustMarshal(&claimsRecord)
	store.Set(addr, bz)
}

// DeleteClaimsRecord deletes a claims record from the store
func (k Keeper) DeleteClaimsRecord(ctx sdk.Context, addr sdk.AccAddress) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixClaimsRecords)
	store.Delete(addr)
}

// IterateClaimsRecords iterates over all claims records and performs a
// callback.
func (k Keeper) IterateClaimsRecords(ctx sdk.Context, handlerFn func(addr sdk.AccAddress, cr types.ClaimsRecord) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixClaimsRecords)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var claimsRecord types.ClaimsRecord
		k.cdc.MustUnmarshal(iterator.Value(), &claimsRecord)

		addr := sdk.AccAddress(iterator.Key()[1:])
		cr := types.ClaimsRecord{
			InitialClaimableAmount: claimsRecord.InitialClaimableAmount,
			ActionsCompleted:       claimsRecord.ActionsCompleted,
		}

		if handlerFn(addr, cr) {
			break
		}
	}
}

// GetClaimsRecords get claims record instances for genesis export
func (k Keeper) GetClaimsRecords(ctx sdk.Context) []types.ClaimsRecordAddress {
	claimsRecords := []types.ClaimsRecordAddress{}
	k.IterateClaimsRecords(ctx, func(addr sdk.AccAddress, cr types.ClaimsRecord) (stop bool) {
		cra := types.ClaimsRecordAddress{
			Address:                addr.String(),
			InitialClaimableAmount: cr.InitialClaimableAmount,
			ActionsCompleted:       cr.ActionsCompleted,
		}

		claimsRecords = append(claimsRecords, cra)
		return false
	})

	return claimsRecords
}
