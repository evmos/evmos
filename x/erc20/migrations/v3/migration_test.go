package v3_test

import (
	"testing"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/evmos/evmos/v12/app"
	"github.com/evmos/evmos/v12/encoding"

	v3types "github.com/evmos/evmos/v12/x/erc20/migrations/v3/types"

	"github.com/evmos/evmos/v12/x/erc20/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/erc20/migrations/v3"
)

type mockSubspace struct {
	ps           v3types.V3Params
	storeKey     storetypes.StoreKey
	transientKey storetypes.StoreKey
}

func newMockSubspace(ps v3types.V3Params, storeKey, transientKey storetypes.StoreKey) mockSubspace {
	return mockSubspace{ps: ps, storeKey: storeKey, transientKey: transientKey}
}

func (ms mockSubspace) GetParamSet(ctx sdk.Context, ps types.LegacyParams) {
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

	var outputParams v3types.V3Params
	inputParams := v3types.DefaultParams()
	legacySubspace := newMockSubspace(v3types.DefaultParams(), storeKey, tKey).WithKeyTable(v3types.ParamKeyTable())
	legacySubspace.SetParamSet(ctx, &inputParams)
	legacySubspace.GetParamSetIfExists(ctx, &outputParams)

	mockSubspace := newMockSubspace(v3types.DefaultParams(), storeKey, tKey)
	require.NoError(t, v3.MigrateStore(ctx, storeKey, mockSubspace))

	// Get all the new parameters from the store
	enableEvmHook := store.Has(types.ParamStoreKeyEnableEVMHook)
	enableErc20 := store.Has(types.ParamStoreKeyEnableErc20)

	params := v3types.NewParams(enableErc20, enableEvmHook)
	require.Equal(t, params, outputParams)
}
