// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
)

// constants
const (
	// module name
	ModuleName = "erc20"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for message routing
	RouterKey = ModuleName
)

// ModuleAddress is the native module address for ERC-20
var ModuleAddress common.Address

func init() {
	ModuleAddress = common.BytesToAddress(authtypes.NewModuleAddress(ModuleName).Bytes())
}

// prefix bytes for the ERC-20 persistent store
const (
	prefixTokenPair = iota + 1
	prefixTokenPairByERC20
	prefixTokenPairByDenom
	prefixSTRv2Addresses
)

// KVStore key prefixes
var (
	KeyPrefixTokenPair        = []byte{prefixTokenPair}
	KeyPrefixTokenPairByERC20 = []byte{prefixTokenPairByERC20}
	KeyPrefixTokenPairByDenom = []byte{prefixTokenPairByDenom}
	KeyPrefixSTRv2Addresses   = []byte{prefixSTRv2Addresses}
)
