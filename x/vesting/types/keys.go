// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

const (
	// ModuleName defines the module's name.
	ModuleName = "vesting"

	// ClawbackKey to be used in the KVStore to track team accounts subject to clawback from governance.
	ClawbackKey = "clawback"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)
