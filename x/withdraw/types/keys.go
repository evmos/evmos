package types

// constants
const (
	// ModuleName defines the withdraw module name
	ModuleName = "withdraw"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for message routing
	RouterKey = ModuleName
)

// prefix bytes for the withdraw module persistent store
const (
	prefixBech32HRP = iota + 1
)

// KVStore key prefixes
var (
	KeyPrefixBech32HRP = []byte{prefixBech32HRP}
)
