package types

// constants
const (
	// module name
	ModuleName = "incentives"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for message routing
	RouterKey = ModuleName
)

// prefix bytes for the EVM persistent store
const (
	prefixIncentive = iota + 1
	prefixGasMeter
)

// KVStore key prefixes
var (
	KeyPrefixIncentive = []byte{prefixIncentive}
	KeyPrefixGasMeter  = []byte{prefixGasMeter}
)

func SplitGasMeterKey(key []byte) (contract string, user string) {
	// TODO
}
