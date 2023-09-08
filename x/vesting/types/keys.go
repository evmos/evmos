// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

const (
	// prefixGovClawbackDisabledKey to be used in the KVStore to track vesting accounts that are not subject
	// to clawback from governance.
	prefixGovClawbackDisabledKey = iota + 1
	// prefixGovClawbackProposalKey to be used in the KVStore to track vesting accounts that are subject
	// to active governance clawback proposals.
	prefixGovClawbackProposalKey
)

var (
	// KeyPrefixGovClawbackDisabledKey is the slice of prefix bytes for storing the governance clawback enabled/disabled flag.
	KeyPrefixGovClawbackDisabledKey = []byte{prefixGovClawbackDisabledKey}
	// KeyPrefixGovClawbackProposalKey is the slice of prefix bytes for storing the vesting account
	// of governance clawback proposals.
	KeyPrefixGovClawbackProposalKey = []byte{prefixGovClawbackProposalKey}
)

const (
	// ModuleName defines the module's name.
	ModuleName = "vesting"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)
