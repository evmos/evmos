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
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/evmos/evmos/v11/x/vesting/types"
)

// Keeper of this module maintains collections of vesting.
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
}

// NewKeeper creates new instances of the vesting Keeper
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	sk types.StakingKeeper,
) Keeper {
	return Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		accountKeeper: ak,
		bankKeeper:    bk,
		stakingKeeper: sk,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// AddVestingAccount adds the address of vesting account to store.
// The caller should check the account type to make sure it's a vesting account type.
func (k Keeper) AddVestingAccount(ctx sdk.Context, addr sdk.AccAddress) {
	// Retrieve the account associated with the address
	acc := k.accountKeeper.GetAccount(ctx, addr)
	if _, ok := acc.(exported.VestingAccount); !ok {
		// Account is not a vesting account
	}
	store := ctx.KVStore(k.storeKey)
	store.Set(types.VestingAccountStoreKey(addr), []byte{0x01})
}

// IterateVestingAccounts iterates over all the stored vesting accounts.
func (k Keeper) IterateVestingAccounts(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixVestingAccounts)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		addr := types.AddressFromVestingAccountKey(iterator.Key())

		acct := k.accountKeeper.GetAccount(ctx, addr)
		if _, ok := acct.(exported.VestingAccount); ok {
			// Account is a vesting account
		} else {
			// Account is not a vesting account
		}

	}
}
