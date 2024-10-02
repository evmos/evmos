package v5_test

import (
	"testing"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/evmos/evmos/v20/encoding"
	v5 "github.com/evmos/evmos/v20/x/feemarket/migrations/v5"
	typesV4 "github.com/evmos/evmos/v20/x/feemarket/migrations/v5/types"
	"github.com/evmos/evmos/v20/x/feemarket/types"

	"github.com/stretchr/testify/require"
)

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeConfig()
	cdc := encCfg.Codec

	storeKey := storetypes.NewKVStoreKey(types.ModuleName)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)

	kvStore := ctx.KVStore(storeKey)

	var v4Params typesV4.ParamsV4
	defaultParams := types.DefaultParams()

	v4Params.NoBaseFee = defaultParams.NoBaseFee
	v4Params.BaseFeeChangeDenominator = defaultParams.BaseFeeChangeDenominator
	v4Params.ElasticityMultiplier = defaultParams.ElasticityMultiplier
	v4Params.EnableHeight = defaultParams.EnableHeight
	v4Params.BaseFee = math.NewInt(1000000)
	v4Params.MinGasPrice = defaultParams.MinGasPrice
	v4Params.MinGasMultiplier = defaultParams.MinGasMultiplier

	v4ParamsBz, err := cdc.Marshal(&v4Params)
	require.NoError(t, err)

	kvStore.Set(types.ParamsKey, v4ParamsBz)

	require.NoError(t, v5.MigrateStore(ctx, storeKey, cdc))

	paramsBz := kvStore.Get(types.ParamsKey)
	var migratedParams types.Params
	cdc.MustUnmarshal(paramsBz, &migratedParams)

	require.Equal(t, v4Params.NoBaseFee, migratedParams.NoBaseFee)
	require.Equal(t, v4Params.BaseFeeChangeDenominator, migratedParams.BaseFeeChangeDenominator)
	require.Equal(t, v4Params.ElasticityMultiplier, migratedParams.ElasticityMultiplier)
	require.Equal(t, v4Params.EnableHeight, migratedParams.EnableHeight)
	require.Equal(t, v4Params.BaseFee, migratedParams.BaseFee.TruncateInt())
	require.Equal(t, v4Params.MinGasPrice, migratedParams.MinGasPrice)
	require.Equal(t, v4Params.MinGasMultiplier, migratedParams.MinGasMultiplier)
}
