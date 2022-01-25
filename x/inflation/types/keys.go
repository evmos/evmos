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

	// module account name for team vesting
	TharsisAccount      = "tharsis_account"
	UnvestedTeamAccount = "unvested_team_account"
)

// ModuleAddress is the native module address for inflation module
var (
	ModuleAddress              common.Address
	TharsisAccountAddress      common.Address
	UnvestedTeamAccountAddress common.Address
)

func init() {
	ModuleAddress = common.BytesToAddress(authtypes.NewModuleAddress(ModuleName).Bytes())
	TharsisAccountAddress = common.BytesToAddress(authtypes.NewModuleAddress(TharsisAccount).Bytes())
	UnvestedTeamAccountAddress = common.BytesToAddress(authtypes.NewModuleAddress(UnvestedTeamAccount).Bytes())
}

// prefix bytes for the inflation persistent store
const (
	prefixPeriod = iota + 1
	prefixEpochMintProvision
)

// KVStore key prefixes
var (
	KeyPrefixPeriod             = []byte{prefixPeriod}
	KeyprefixEpochMintProvision = []byte{prefixEpochMintProvision}
)
