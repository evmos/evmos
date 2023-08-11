// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

const (
	// prefixGovClawbackDisabledKey to be used in the KVStore to track vesting accounts that are not subject
	// to clawback from governance.
	prefixGovClawbackDisabledKey = iota + 1
)

// KeyPrefixGovClawbackDisabledKey is the prefix bytes for storing the governance clawback enabled/disabled flag.
var KeyPrefixGovClawbackDisabledKey = []byte{prefixGovClawbackDisabledKey}

const (
	// ModuleName defines the module's name.
	ModuleName = "vesting"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)
