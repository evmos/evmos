package v3_test

import (
	"testing"

	"github.com/evmos/evmos/v10/x/erc20/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v10/x/erc20/migrations/v3"
	v3types "github.com/evmos/evmos/v10/x/erc20/migrations/v3/types"

	"github.com/evmos/ethermint/encoding"

	"github.com/evmos/evmos/v10/app"
)

type mockSubspace struct {
	ps types.Params
}

func newMockSubspace(ps types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSet(ctx sdk.Context, ps types.LegacyParams) {
	*ps.(*types.Params) = ms.ps
}

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	legacySubspace := newMockSubspace(types.DefaultParams())
	require.NoError(t, v3.MigrateStore(ctx, store, legacySubspace, cdc))

	// Get all the new parameters from the store
	var enableEvmHook gogotypes.BoolValue
	bz := store.Get(v3types.ParamStoreKeyEnableEVMHook)
	cdc.MustUnmarshal(bz, &enableEvmHook)

	var enableErc20 gogotypes.BoolValue
	bz = store.Get(v3types.ParamStoreKeyEnableErc20)
	cdc.MustUnmarshal(bz, &enableErc20)

	params := types.NewParams(enableErc20.Value, enableEvmHook.Value)
	require.Equal(t, legacySubspace.ps, params)
}
