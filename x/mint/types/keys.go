package types

// MinterKey is the key to use for the keeper store at which
// the Minter and its DailyProvisions are stored.
var MinterKey = []byte{0x00}

// LastReductionTimeKey is the key to use for the keeper store
// for sacreng the last time at which reduction occurred.
var LastReductionTimeKey = []byte{0x01}

const (
	// ModuleName is the module name.
	ModuleName = "mint"

	// StoreKey is the default store key for mint.
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the minting store.
	QuerierRoute = StoreKey
)
