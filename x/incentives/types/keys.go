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
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
)

// constants
const (
	// module name
	ModuleName = "incentives"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for message routing
	RouterKey = ModuleName
)

// ModuleAddress is the native module address for incentives module
var ModuleAddress common.Address

func init() {
	ModuleAddress = common.BytesToAddress(authtypes.NewModuleAddress(ModuleName).Bytes())
}

// prefix bytes for the incentives persistent store
const (
	prefixIncentive = iota + 1
	prefixGasMeter
	prefixAllocationMeter
)

// KVStore key prefixes
var (
	KeyPrefixIncentive       = []byte{prefixIncentive}
	KeyPrefixGasMeter        = []byte{prefixGasMeter}
	KeyPrefixAllocationMeter = []byte{prefixAllocationMeter}
)

// SplitGasMeterKey is a helper to split up KV-store keys in a
// `prefix|<contract_address>|<participant_address>` format
func SplitGasMeterKey(key []byte) (contract, userAddr common.Address) {
	// with prefix
	if len(key) == 41 {
		key = key[1:]
	}

	contract = common.BytesToAddress(key[:common.AddressLength])
	userAddr = common.BytesToAddress(key[common.AddressLength:])
	return contract, userAddr
}
