package v4_test

import (
	"testing"

	v4 "github.com/evmos/evmos/v18/x/erc20/migrations/v4"
	v4types "github.com/evmos/evmos/v18/x/erc20/migrations/v4/types"

	"github.com/evmos/evmos/v18/x/erc20/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var isTrue = []byte("0x01")

func TestMigrate(t *testing.T) {
	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	store.Set(v4types.ParamStoreKeyEnableEVMHook, isTrue)
	store.Set(v4types.ParamStoreKeyEnableErc20, isTrue)

	require.NoError(t, v4.MigrateStore(ctx, storeKey))

	// Get all the new parameters from the store
	enableEvmHook := store.Has(v4types.ParamStoreKeyEnableEVMHook)
	enableErc20 := store.Has(v4types.ParamStoreKeyEnableErc20)

	dynamicbz := store.Get(types.ParamStoreKeyDynamicPrecompiles)
	nativebz := store.Get(types.ParamStoreKeyNativePrecompiles)

	require.ElementsMatch(t, dynamicbz, []byte{})
	require.ElementsMatch(t, nativebz, []byte{})
	require.False(t, enableEvmHook, "params should have been deleted")
	require.True(t, enableErc20, "params should be enabled")
}
