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

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

const (
	// ModuleName defines the module's name.
	ModuleName = "vesting"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

const (
	prefixVestingAccounts = iota + 1
)

var (
	KeyPrefixVestingAccounts = []byte{prefixVestingAccounts}
)

// VestingAccountStoreKey turn an address to key used to record it in the vesting store
func VestingAccountStoreKey(addr sdk.AccAddress) []byte {
	return append(KeyPrefixVestingAccounts, addr.Bytes()...)
}

// AddressFromVestingAccountKey creates the address from VestingAccountKey
func AddressFromVestingAccountKey(key []byte) sdk.AccAddress {
	kv.AssertKeyAtLeastLength(key, 1)
	return key[1:] // remove prefix byte
}
