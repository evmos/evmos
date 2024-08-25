// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

const (
	// ModuleName defines the module name
	ModuleName = "staterent"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for staterent
	RouterKey = ModuleName
)

// prefix bytes
const (
	prefixPrefixParams = iota + 1
	prefixFlaggedInfo
)

// KVStore key prefixes
var (
	KeyPrefixParams      = []byte{prefixPrefixParams}
	KeyPrefixFlaggedInfo = []byte{prefixFlaggedInfo}
)
