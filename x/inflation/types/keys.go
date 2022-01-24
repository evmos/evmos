package types

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
)

// constants
const (
	// module name
	ModuleName = "inflation"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for message routing
	RouterKey = ModuleName
)

// ModuleAddress is the native module address for inflation module
var ModuleAddress common.Address

func init() {
	ModuleAddress = common.BytesToAddress(authtypes.NewModuleAddress(ModuleName).Bytes())
}

// prefix bytes for the inflation persistent store
const (
	prefixPeriod = iota + 1
	prefixEpochMintProvisions
)

// KVStore key prefixes
var (
	KeyPrefixPeriod              = []byte{prefixPeriod}
	KeyprefixEpochMintProvisions = []byte{prefixEpochMintProvisions}
)

// module account name for team vesting
const TharsisAccount = "tharsis_account"
