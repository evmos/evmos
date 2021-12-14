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
// Change StoreKey from `intrarelayer/` to `erc20/` for:
// - TokenPair
// - TokenPairByERC20 (Change TokenPairByERC20 Value)
// - TokenPairByDenom (Change TokenPairByDenom Prefix + Value)
// - Params
// - Bank balances
func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	migrateTokenPairKeys(store, store, cdc)
	migrateParameterKeys(store, store, cdc)
	// migrateBalanceKeys()
}

// migrateTokenPairKeys migrates TokenPair keys to use new Storekey
func migrateTokenPairKeys(irmStore, erc20Store sdk.KVStore, cdc codec.BinaryCodec) error {
	// old key is of format: `intrarelayer` || key prefix || id bytes
	// new key is of format: `erc20` || key prefix | id bytes
	keyPrefixTokenPair := types.KeyPrefixTokenPair
	oldStore := prefix.NewStore(irmStore, keyPrefixTokenPair)
	newStore := prefix.NewStore(erc20Store, keyPrefixTokenPair)

	// Old and new TokenPairByERC20 store
	keyPrefixTokenPairByERC20 := types.KeyPrefixTokenPairByERC20
	oldStoreTokenPairByERC20 := prefix.NewStore(irmStore, keyPrefixTokenPairByERC20)
	newStoreTokenPairByERC20 := prefix.NewStore(erc20Store, keyPrefixTokenPairByERC20)

	// Old and new TokenPairByDenom store
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

		if strings.HasPrefix(tokenPair.Denom, "intrarelayer") {
			// Migrate TokenPair key and value with updated TokenPair denom
			oldErc20 := tokenPair.GetERC20Contract()
			oldDenomKey := tokenPair.Denom
			tokenPair.Denom = strings.ReplaceAll(tokenPair.Denom, "intrarelayer", "erc20")
			bz = cdc.MustMarshal(&tokenPair)
			newStore.Set([]byte(id), bz)
			oldStore.Delete(key)

			// Migrate TokenPairByErc20 value (erc20 Prefix => newID)
			erc20 := tokenPair.GetERC20Contract()
			erc20Key := append(types.KeyPrefixTokenPairByERC20, erc20.Bytes()...)
			irmKey := append(types.KeyPrefixTokenPairByERC20, oldErc20.Bytes()...)
			newStoreTokenPairByERC20.Set(erc20Key, []byte(id))
			oldStoreTokenPairByERC20.Delete(irmKey)

			// Migrate TokenPairByDenom key and value (newID => newTokenPair)
			denom := tokenPair.Denom
			denomKey := append(types.KeyPrefixTokenPairByDenom, []byte(denom)...)
			newStoreTokenPairByDenom.Set(denomKey, []byte(id))
			oldStoreTokenPairByDenom.Delete([]byte(oldDenomKey))
		}
	}
	return nil
}

// migrateTokenPairKeys migrates Parameter keys to use new Storekey
func migrateParameterKeys(irmStore, erc20Store sdk.KVStore, cdc codec.BinaryCodec) error {
	// old key is of format: `intrarelayer` || paramStoreKey
	// new key is of format: `erc20` || paramStoreKey
	paramEnableErc20 := types.ParamStoreKeyEnableErc20
	oldStore := prefix.NewStore(irmStore, paramEnableErc20)
	newStore := prefix.NewStore(erc20Store, paramEnableErc20)

	var enableErc20 bool
	bz := oldStore.Get(paramEnableErc20)
	cdc.MustUnmarshal(bz, &enableErc20)
	// store param in new store

	return nil
}
