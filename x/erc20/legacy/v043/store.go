package v043

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MigrateStore performs in-place store migrations from v0.42 to v0.43. The
// migration includes:
//
// - Keys from the store need to be migrated to the new module
// - Change addresses to be length-prefixed.
// - Change balances prefix to 1 byte
// - Change supply to be indexed by denom
func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryMarshaler) error {
	store := ctx.KVStore(storeKey)

	migrateBalanceKeys(store)
	return migrateSupply(store, cdc)
}
