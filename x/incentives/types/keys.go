package types

import (
	"bytes"

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

func SplitGasMeterKey(key []byte) (contract, userAddr common.Address) {
	keySplit := bytes.SplitN(key, []byte(""), 41)
	k1 := bytes.Join(keySplit[1:common.AddressLength+1], []byte(""))
	k2 := bytes.Join(keySplit[common.AddressLength+1:], []byte(""))
	contract = common.BytesToAddress(k1)
	userAddr = common.BytesToAddress(k2)
	return
}
