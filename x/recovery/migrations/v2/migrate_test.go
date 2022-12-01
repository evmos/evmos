package v2_test

import (
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/encoding"
	"github.com/evmos/evmos/v10/app"
	v2 "github.com/evmos/evmos/v10/x/recovery/migrations/v2"
	v2types "github.com/evmos/evmos/v10/x/recovery/migrations/v2/types"
	"github.com/evmos/evmos/v10/x/recovery/types"
	"github.com/stretchr/testify/require"
	"testing"
)

type mockSubspace struct {
	ps v2types.Params
}

func newMockSubspace(ps v2types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSetIfExists(ctx sdk.Context, ps types.LegacyParams) {
	*ps.(*v2types.Params) = ms.ps
}

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	kvStore := ctx.KVStore(storeKey)

	legacySubspace := newMockSubspace(v2types.DefaultParams())
	require.NoError(t, v2.MigrateStore(ctx, storeKey, legacySubspace, cdc))

	paramsBz := kvStore.Get(v2types.ParamsKey)
	var params v2types.Params
	cdc.MustUnmarshal(paramsBz, &params)

	require.Equal(t, params, legacySubspace.ps)
}
