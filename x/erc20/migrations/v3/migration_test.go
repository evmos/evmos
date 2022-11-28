package v3_test

import (
	"bytes"
	"testing"

	"github.com/evmos/evmos/v10/x/erc20/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/encoding"
	"github.com/evmos/evmos/v10/x/erc20/migrations/v3"

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
	var enableEvmHook bool
	bz := store.Get(types.ParamStoreKeyEnableEVMHook)
	if bytes.Equal(bz, []byte("0x01")) {
		enableEvmHook = true
	}

	var enableErc20 bool
	bz = store.Get(types.ParamStoreKeyEnableErc20)
	if bytes.Equal(bz, []byte("0x01")) {
		enableErc20 = true
	}

	params := types.NewParams(enableErc20, enableEvmHook)
	require.Equal(t, legacySubspace.ps, params)
}
