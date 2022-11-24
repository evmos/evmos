package v2_test

import (
	"github.com/evmos/evmos/v10/x/inflation/exported"
	v2 "github.com/evmos/evmos/v10/x/inflation/migrations/v2"
	"github.com/evmos/evmos/v10/x/inflation/types"
	gogotypes "github.com/gogo/protobuf/types"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/app"
	"github.com/evmos/ethermint/encoding"
	v2types "github.com/evmos/evmos/v10/x/inflation/migrations/v2/types"
	"github.com/stretchr/testify/require"
)

type mockSubspace struct {
	ps v2types.Params
}

func newMockSubspace(ps v2types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSet(ctx sdk.Context, ps exported.Params) {
	*ps.(*v2types.Params) = ms.ps
}

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	legacySubspace := newMockSubspace(v2types.DefaultParams())
	require.NoError(t, v2.MigrateStore(ctx, store, legacySubspace, cdc))

	// Get all the new parameters from the store
	var mintDenom gogotypes.StringValue
	bz := store.Get(v2types.ParamStoreKeyMintDenom)
	cdc.MustUnmarshal(bz, &mintDenom)

	var enableInflation gogotypes.BoolValue
	bz = store.Get(v2types.ParamStoreKeyEnableInflation)
	cdc.MustUnmarshal(bz, &enableInflation)

	var inflationDistribution v2types.InflationDistribution
	bz = store.Get(v2types.ParamStoreKeyInflationDistribution)
	cdc.MustUnmarshal(bz, &inflationDistribution)

	var exponentialCalculation v2types.ExponentialCalculation
	bz = store.Get(v2types.ParamStoreKeyExponentialCalculation)
	cdc.MustUnmarshal(bz, &exponentialCalculation)

	params := v2types.NewParams(mintDenom.Value, exponentialCalculation, inflationDistribution, enableInflation.Value)
	require.Equal(t, legacySubspace.ps, params)
}
