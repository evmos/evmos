package v5

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/x/erc20/types"
)

// MigrateStore migrates the x/erc20 module state from the consensus version 4 to
// version 5. Specifically, it takes the token pairs and stores them in the new format.
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefixTokenPair)
	defer iterator.Close()

	var err error
	for ; iterator.Valid(); iterator.Next() {
		var tokenPair types.TokenPair
		err = tokenPair.Unmarshal(iterator.Value())
		if err != nil {
			return err
		}

		tokenPair.ContractOwnerAddress = ""
		marshaledPair, err := tokenPair.Marshal()
		if err != nil {
			return err
		}
		store.Set(tokenPair.GetID(), marshaledPair)
	}

	return nil
}
