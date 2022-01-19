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

// ModuleAddress is the native module address for incentives module
var ModuleAddress common.Address

func init() {
	ModuleAddress = common.BytesToAddress(authtypes.NewModuleAddress(ModuleName).Bytes())
}

// prefix bytes for the incentives persistent store
const (
	prefixMinter = iota + 1
)

// KVStore key prefixes
var (
	KeyPrefixMinter = []byte{prefixMinter}
)

// ___________________________________________________________________________

// TODO Refactor OSMOSIS mint Keys

// MinterKey is the key to use for the keeper store.
var MinterKey = []byte{0x00}

// LastHalvenEpochKey is the key to use for the keeper store.
var LastHalvenEpochKey = []byte{0x03}

const (
	// module acct name for developer vesting
	DeveloperVestingModuleAcctName = "developer_vesting_unvested"

	// QuerierRoute is the querier route for the minting store.
	QuerierRoute = StoreKey

	// Query endpoints supported by the minting querier
	QueryParameters      = "parameters"
	QueryEpochProvisions = "epoch_provisions"
)
