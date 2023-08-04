package v3_test

import (
	"testing"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v14/app"
	"github.com/evmos/evmos/v14/encoding"
	v3 "github.com/evmos/evmos/v14/x/claims/migrations/v3"
	v3types "github.com/evmos/evmos/v14/x/claims/migrations/v3/types"
	"github.com/evmos/evmos/v14/x/claims/types"
	"github.com/stretchr/testify/require"
)

type mockSubspace struct {
	ps           v3types.V3Params
	storeKey     storetypes.StoreKey
	transientKey storetypes.StoreKey
}

func newMockSubspace(ps v3types.V3Params, storeKey, transientKey storetypes.StoreKey) mockSubspace {
	return mockSubspace{ps: ps, storeKey: storeKey, transientKey: transientKey}
}

func (ms mockSubspace) GetParamSetIfExists(_ sdk.Context, ps types.LegacyParams) {
	*ps.(*v3types.V3Params) = ms.ps
}

func (ms mockSubspace) WithKeyTable(keyTable paramtypes.KeyTable) paramtypes.Subspace {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec
	return paramtypes.NewSubspace(cdc, encCfg.Amino, ms.storeKey, ms.transientKey, "test").WithKeyTable(keyTable)
}

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	kvStore := ctx.KVStore(storeKey)

	var outputParams v3types.V3Params
	inputParams := v3types.DefaultParams()
	legacySubspace := newMockSubspace(v3types.DefaultParams(), storeKey, tKey).WithKeyTable(v3types.ParamKeyTable())
	legacySubspace.SetParamSet(ctx, &inputParams)
	legacySubspace.GetParamSetIfExists(ctx, &outputParams)

	mockSubspace := newMockSubspace(v3types.DefaultParams(), storeKey, tKey)
	require.NoError(t, v3.MigrateStore(ctx, storeKey, mockSubspace, cdc))

	paramsBz := kvStore.Get(v3types.ParamsKey)
	var params v3types.V3Params
	cdc.MustUnmarshal(paramsBz, &params)

	require.Equal(t, params, outputParams)
}
