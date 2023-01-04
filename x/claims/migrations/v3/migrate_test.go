package v3_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/encoding"
	"github.com/evmos/evmos/v10/app"
	v3 "github.com/evmos/evmos/v10/x/claims/migrations/v3"
	v3types "github.com/evmos/evmos/v10/x/claims/migrations/v3/types"
	"github.com/evmos/evmos/v10/x/claims/types"
	"github.com/stretchr/testify/require"
)

type mockSubspace struct {
	ps v3types.V3Params
}

func newMockSubspace(ps v3types.V3Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSetIfExists(ctx sdk.Context, ps types.LegacyParams) {
	*ps.(*v3types.V3Params) = ms.ps
}

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	kvStore := ctx.KVStore(storeKey)

	legacySubspace := newMockSubspace(v3types.DefaultParams())
	require.NoError(t, v3.MigrateStore(ctx, storeKey, legacySubspace, cdc))

	paramsBz := kvStore.Get(v3types.ParamsKey)
	var params v3types.V3Params
	cdc.MustUnmarshal(paramsBz, &params)

	require.Equal(t, params, legacySubspace.ps)
}
