package v05

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/evmos/x/erc20/types"
)

// MigrateStore performs in-place store migrations from v0.42 to v0.43. The
// migration includes:
//
// - Change StoreKey from `intrarelayer/` to `erc20/` for:
// 		- TokenPair
// 		- TokenPairByERC20
// 		- TokenPairByDenom
// - Change TokenPairByERC20 Value
// - Change TokenPairByDenom Prefix + Value
func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	migrateTokenPairKeys(store, store, cdc)
}

// migrateTokenPairKeys migrates TokenPair keys to use new Storekey
func migrateTokenPairKeys(irmStore, erc20Store sdk.KVStore, cdc codec.BinaryCodec) error {
	// old key is of format: `intrarelayer` || key prefix || id bytes
	// new key is of format: `erc20` || key prefix | id bytes
	keyPrefixTokenPair := types.KeyPrefixTokenPair
	oldStore := prefix.NewStore(irmStore, keyPrefixTokenPair)
	newStore := prefix.NewStore(erc20Store, keyPrefixTokenPair)

	keyPrefixTokenPairByERC20 := types.KeyPrefixTokenPairByERC20
	storeTokenPairByERC20 := prefix.NewStore(irmStore, keyPrefixTokenPairByERC20)

	keyPrefixTokenPairByDenom := types.KeyPrefixTokenPairByDenom
	oldStoreTokenPairByDenom := prefix.NewStore(irmStore, keyPrefixTokenPairByDenom)
	newStoreTokenPairByDenom := prefix.NewStore(erc20Store, keyPrefixTokenPairByDenom)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		key := oldStoreIter.Key()

		// Get TokenPair from old store
		var tokenPair types.TokenPair
		bz := oldStore.Get(key)
		cdc.MustUnmarshal(bz, &tokenPair)
		id := tokenPair.GetID()

		// Update TokenPair denom and write new id to newStore
		if strings.HasPrefix(tokenPair.Denom, "intrarelayer") {
			oldDenomKey := tokenPair.Denom
			tokenPair.Denom = strings.ReplaceAll(tokenPair.Denom, "intrarelayer", "erc20")
			bz = cdc.MustMarshal(&tokenPair)
			// Set new key on store
			newStore.Set([]byte(id), bz)
			oldStore.Delete(key)

			// TODO: migrate TokenPairByErc20 value (id)
			// get Tokenpair erc20 key
			erc20 := tokenPair.GetERC20Contract()
			erc20Key := types.KeyPrefixTokenPairByERC20 + []byte(erc20)
			// set erc20 Prefix => newID to new store
			storeTokenPairByERC20.Set(erc20Key, []byte(id))

			// TODO: migrate TokenPairByDenom key and value (newID => newTokenPair)
			// get newDenom Prefix
			denom := tokenPair.Denom
			denomKey := types.KeyPrefixTokenPairByDenom + []byte(denom)
			// set newDenom => newID to new store
			newStoreTokenPairByDenom.Set(denomKey, []byte(id))
			oldStoreTokenPairByDenom.Delete([]byte(oldDenomKey))
		}
	}
	return nil
}
