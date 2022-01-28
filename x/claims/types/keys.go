package types

const (
	// ModuleName defines the module name
	ModuleName = "claims"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for claims
	RouterKey = ModuleName
)

// prefix bytes for the claims module's persistent store
const (
	prefixClaimsRecords = iota + 1
)

// KVStore key prefixes
var (
	KeyPrefixClaimsRecords = []byte{prefixClaimsRecords}
)
