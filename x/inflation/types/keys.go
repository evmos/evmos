package types

// constants
const (
	// module name
	ModuleName = "inflation"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for message routing
	RouterKey = ModuleName

	// module account name for team vesting
	UnvestedTeamAccount = "unvested_team_account"
)

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
