package v3_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	v3 "github.com/evmos/evmos/v19/x/inflation/v1/migrations/v3"
	"github.com/evmos/evmos/v19/x/inflation/v1/types"
	"github.com/stretchr/testify/require"
)

func TestMigrate(t *testing.T) {
	storeKey := storetypes.NewKVStoreKey(types.ModuleName)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	store.Set(v3.KeyPrefixEpochMintProvision, []byte{0x01})
	epochMintProvision := store.Get(v3.KeyPrefixEpochMintProvision)
	require.Equal(t, epochMintProvision, []byte{0x01})

	require.NoError(t, v3.MigrateStore(store))
	require.False(t, store.Has(v3.KeyPrefixEpochMintProvision))
}
