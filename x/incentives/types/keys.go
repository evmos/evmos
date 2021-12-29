package types

import (
	"github.com/ethereum/go-ethereum/common"
)

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

// SplitGasMeterKey is a helper to split up KV-store keys in a
// `prefix-contract-participant` format
func SplitGasMeterKey(key []byte) (contract, userAddr common.Address) {
	// with prefix
	if len(key) == 41 {
		key = key[1:]
	}

	contract = common.BytesToAddress(key[:common.AddressLength])
	userAddr = common.BytesToAddress(key[common.AddressLength:])
	return contract, userAddr
}
