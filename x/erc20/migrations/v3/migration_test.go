package v3_test

import (
	"testing"

	v3types "github.com/evmos/evmos/v10/x/erc20/migrations/v3/types"

	"github.com/evmos/evmos/v10/x/erc20/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v10/x/erc20/migrations/v3"
)

type mockSubspace struct {
	ps v3types.V3Params
}

func newMockSubspace(ps v3types.V3Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSet(ctx sdk.Context, ps types.LegacyParams) {
	*ps.(*v3types.V3Params) = ms.ps
}

func TestMigrate(t *testing.T) {
	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	legacySubspace := newMockSubspace(v3types.DefaultParams())
	require.NoError(t, v3.MigrateStore(ctx, storeKey, legacySubspace))

	// Get all the new parameters from the store
	enableEvmHook := store.Has(types.ParamStoreKeyEnableEVMHook)
	enableErc20 := store.Has(types.ParamStoreKeyEnableErc20)

	params := v3types.NewParams(enableErc20, enableEvmHook)
	require.Equal(t, legacySubspace.ps, params)
}
