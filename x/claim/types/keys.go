package types

const (
	// ModuleName defines the module name
	ModuleName = "claim"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName
)

// prefix bytes for the claim module's persistent store
const (
	prefixClaimRecords = iota + 1
)

// KVStore key prefixes
var (
	KeyPrefixClaimRecords = []byte{prefixClaimRecords}
)
