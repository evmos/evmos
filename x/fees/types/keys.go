package types

// constants
const (
	// module name
	ModuleName = "fees"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for message routing
	RouterKey = ModuleName
)

// prefix bytes for the fees persistent store
const (
	prefixFee = iota + 1
	prefixFeeWithdrawal
	prefixFeeInverse
)

// KVStore key prefixes
var (
	KeyPrefixFee           = []byte{prefixFee}
	KeyPrefixFeeWithdrawal = []byte{prefixFeeWithdrawal}
	KeyPrefixInverse       = []byte{prefixFeeInverse}
)
