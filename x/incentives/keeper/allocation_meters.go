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
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v10/x/incentives/types"
)

// GetAllAllocationMeters - get all registered AllocationMeters
func (k Keeper) GetAllAllocationMeters(ctx sdk.Context) []sdk.DecCoin {
	allocationMeters := []sdk.DecCoin{}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixAllocationMeter)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		denom := string(iterator.Key()[1:])
		allocationMeter, found := k.GetAllocationMeter(ctx, denom)
		if !found {
			continue
		}

		allocationMeters = append(allocationMeters, allocationMeter)
	}

	return allocationMeters
}

// GetAllocationMeter - get registered allocationMeter from the identifier
func (k Keeper) GetAllocationMeter(
	ctx sdk.Context,
	denom string,
) (sdk.DecCoin, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllocationMeter)

	bz := store.Get([]byte(denom))
	if bz == nil {
		return sdk.DecCoin{
			Denom:  denom,
			Amount: sdk.ZeroDec(),
		}, false
	}

	var amount sdk.Dec
	err := amount.Unmarshal(bz)
	if err != nil {
		panic(fmt.Errorf("unable to unmarshal amount value %v", err))
	}
	return sdk.DecCoin{
		Denom:  denom,
		Amount: amount,
	}, true
}

// SetAllocationMeter stores an allocationMeter
func (k Keeper) SetAllocationMeter(ctx sdk.Context, am sdk.DecCoin) {
	bz, err := am.Amount.Marshal()
	if err != nil {
		panic(fmt.Errorf("unable to marshal amount value %v", err))
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllocationMeter)
	key := []byte(am.Denom)

	// Remove zero allocation meters
	if am.IsZero() {
		store.Delete(key)
	} else {
		store.Set(key, bz)
	}
}
