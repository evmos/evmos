package keeper_test

import (
	"reflect"
	"testing"

	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v19/x/feemarket/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	nw := network.NewUnitTestNetwork()
	ctx := nw.GetContext()

	params := nw.App.FeeMarketKeeper.GetParams(ctx)
	require.NotNil(t, params.BaseFee)
	require.NotNil(t, params.MinGasPrice)
	require.NotNil(t, params.MinGasMultiplier)
}

func TestSetGetParams(t *testing.T) {
	nw := network.NewUnitTestNetwork()
	ctx := nw.GetContext()

	params := types.DefaultParams()
	err := nw.App.FeeMarketKeeper.SetParams(ctx, params)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		paramsFun func() interface{}
		getFun    func() interface{}
		expected  bool
	}{
		{
			"success - Checks if the default params are set correctly",
			func() interface{} {
				return types.DefaultParams()
			},
			func() interface{} {
				return nw.App.FeeMarketKeeper.GetParams(ctx)
			},
			true,
		},
		{
			"success - Check ElasticityMultiplier is set to 3 and can be retrieved correctly",
			func() interface{} {
				params.ElasticityMultiplier = 3
				err := nw.App.FeeMarketKeeper.SetParams(ctx, params)
				require.NoError(t, err)
				return params.ElasticityMultiplier
			},
			func() interface{} {
				return nw.App.FeeMarketKeeper.GetParams(ctx).ElasticityMultiplier
			},
			true,
		},
		{
			"success - Check BaseFeeEnabled is computed with its default params and can be retrieved correctly",
			func() interface{} {
				err := nw.App.FeeMarketKeeper.SetParams(ctx, types.DefaultParams())
				require.NoError(t, err)
				return true
			},
			func() interface{} {
				return nw.App.FeeMarketKeeper.GetBaseFeeEnabled(ctx)
			},
			true,
		},
		{
			"success - Check BaseFeeEnabled is computed with alternate params and can be retrieved correctly",
			func() interface{} {
				params.NoBaseFee = true
				params.EnableHeight = 5
				err := nw.App.FeeMarketKeeper.SetParams(ctx, params)
				require.NoError(t, err)
				return true
			},
			func() interface{} {
				return nw.App.FeeMarketKeeper.GetBaseFeeEnabled(ctx)
			},
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outcome := reflect.DeepEqual(tc.paramsFun(), tc.getFun())
			require.Equal(t, tc.expected, outcome)
		})
	}
}
