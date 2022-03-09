package types

// constants
const (
	// module name
	ModuleName = "inflation"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for message routing
	RouterKey = ModuleName
)

// prefix bytes for the inflation persistent store
const (
	prefixPeriod = iota + 1
	prefixEpochMintProvision
	prefixEpochIdentifier
	prefixEpochsPerPeriod
	prefixSkippedEpochs
)

// KVStore key prefixes
var (
	KeyPrefixPeriod             = []byte{prefixPeriod}
	KeyPrefixEpochMintProvision = []byte{prefixEpochMintProvision}
	KeyPrefixEpochIdentifier    = []byte{prefixEpochIdentifier}
	KeyPrefixEpochsPerPeriod    = []byte{prefixEpochsPerPeriod}
	KeyPrefixSkippedEpochs      = []byte{prefixSkippedEpochs}
)
