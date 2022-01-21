package types

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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

// ModuleAddress is the native module address for incentives module
var ModuleAddress common.Address

func init() {
	ModuleAddress = common.BytesToAddress(authtypes.NewModuleAddress(ModuleName).Bytes())
}

// prefix bytes for the incentives persistent store
const (
	prefixIncentive = iota + 1
	prefixGasMeter
	prefixAllocationMeter
)

// KVStore key prefixes
var (
	KeyPrefixIncentive       = []byte{prefixIncentive}
	KeyPrefixGasMeter        = []byte{prefixGasMeter}
	KeyPrefixAllocationMeter = []byte{prefixAllocationMeter}
)

// SplitGasMeterKey is a helper to split up KV-store keys in a
// `prefix|<contract_address>|<participant_address>` format
func SplitGasMeterKey(key []byte) (contract, userAddr common.Address) {
	// with prefix
	if len(key) == 41 {
		key = key[1:]
	}

	contract = common.BytesToAddress(key[:common.AddressLength])
	userAddr = common.BytesToAddress(key[common.AddressLength:])
	return contract, userAddr
}
