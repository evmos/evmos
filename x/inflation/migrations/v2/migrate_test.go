package v2_test

import (
	"testing"

	v2 "github.com/evmos/evmos/v10/x/inflation/migrations/v2"
	"github.com/evmos/evmos/v10/x/inflation/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/app"
	"github.com/evmos/ethermint/encoding"
	v2types "github.com/evmos/evmos/v10/x/inflation/migrations/v2/types"
	"github.com/stretchr/testify/require"
)

type mockSubspace struct {
	ps v2types.V2Params
}

func newMockSubspaceEmpty() mockSubspace {
	return mockSubspace{}
}

func newMockSubspace(ps v2types.V2Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSetIfExists(ctx sdk.Context, ps types.LegacyParams) {
	*ps.(*v2types.V2Params) = ms.ps
}

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	legacySubspace := newMockSubspace(v2types.DefaultParams())
	require.NoError(t, v2.MigrateStore(ctx, storeKey, legacySubspace, cdc))

	legacySubspaceEmpty := newMockSubspaceEmpty()
	require.Error(t, v2.MigrateStore(ctx, storeKey, legacySubspaceEmpty, cdc))

	var params v2types.V2Params
	paramsBz := store.Get(v2types.ParamsKey)
	cdc.MustUnmarshal(paramsBz, &params)

	require.Equal(t, params, legacySubspace.ps)
}
