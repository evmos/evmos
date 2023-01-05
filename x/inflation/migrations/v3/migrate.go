package v3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// prefix bytes for the inflation persistent store
const prefixEpochMintProvision = 2

// KeyPrefixEpochMintProvision key prefix
var KeyPrefixEpochMintProvision = []byte{prefixEpochMintProvision}

// MigrateStore migrates the x/inflation module state from the consensus version 2 to
// version 3. Specifically, it deletes the EpochMintProvision from the store
func MigrateStore(store sdk.KVStore) error {
	store.Delete(KeyPrefixEpochMintProvision)
	return nil
}
