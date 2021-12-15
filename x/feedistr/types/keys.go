package types

// constants
const (
	// module name
	ModuleName = "distribution"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for message routing
	RouterKey = ModuleName
)

// prefix bytes for the EVM persistent store
const (
	prefixContractOwner = iota + 1
	prefixContractOwnerInverse
)

// KVStore key prefixes
var (
	KeyPrefixContractOwner        = []byte{prefixContractOwner}
	KeyPrefixContractOwnerInverse = []byte{prefixContractOwnerInverse}
)
