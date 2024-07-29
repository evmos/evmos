package v4_test

import (
	"testing"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/evmos/evmos/v19/app"
	"github.com/evmos/evmos/v19/encoding"
	"github.com/stretchr/testify/require"

	v3types "github.com/evmos/evmos/v19/x/erc20/migrations/v3/types"
	v4 "github.com/evmos/evmos/v19/x/erc20/migrations/v4"

	"github.com/evmos/evmos/v19/x/erc20/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type mockSubspace struct {
	ps           v3types.V3Params
	storeKey     storetypes.StoreKey
	transientKey storetypes.StoreKey
}

func newMockSubspace(ps v3types.V3Params, storeKey, transientKey storetypes.StoreKey) mockSubspace {
	return mockSubspace{ps: ps, storeKey: storeKey, transientKey: transientKey}
}

func (ms mockSubspace) GetParamSet(_ sdk.Context, ps types.LegacyParams) {
	*ps.(*v3types.V3Params) = ms.ps
}

func (ms mockSubspace) WithKeyTable(keyTable paramtypes.KeyTable) paramtypes.Subspace {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec
	return paramtypes.NewSubspace(cdc, encCfg.Amino, ms.storeKey, ms.transientKey, "test").WithKeyTable(keyTable)
}

func TestMigrate(t *testing.T) {
	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	inputParams := v3types.DefaultParams()
	legacySubspace := newMockSubspace(v3types.DefaultParams(), storeKey, tKey).WithKeyTable(v3types.ParamKeyTable())
	legacySubspace.SetParamSet(ctx, &inputParams)

	require.NoError(t, v4.MigrateStore(ctx, storeKey))

	// Get all the new parameters from the store
	enableEvmHook := store.Has(v3types.ParamStoreKeyEnableEVMHook)
	require.False(t, enableEvmHook)

	enableErc20 := store.Has(types.ParamStoreKeyEnableErc20)
	require.True(t, enableErc20)

	var dynamicPrecompiles []string
	bz := store.Get(types.ParamStoreKeyDynamicPrecompiles)
	for i := 0; i < len(bz); i += v4.AddressLength {
		dynamicPrecompiles = append(dynamicPrecompiles, string(bz[i:i+v4.AddressLength]))
	}

	var nativePrecompiles []string
	bz = store.Get(types.ParamStoreKeyNativePrecompiles)
	for i := 0; i < len(bz); i += v4.AddressLength {
		nativePrecompiles = append(nativePrecompiles, string(bz[i:i+v4.AddressLength]))
	}

	params := types.NewParams(enableErc20, nativePrecompiles, dynamicPrecompiles)
	defaultParams := types.DefaultParams()
	require.Equal(t, params, defaultParams)
}
