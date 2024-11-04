package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v20/x/feemarket/types"
	"github.com/stretchr/testify/require"
)

func TestQueryParams(t *testing.T) {
	var (
		nw  *network.UnitTestNetwork
		ctx sdk.Context
	)

	testCases := []struct {
		name    string
		expPass bool
	}{
		{
			"pass",
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// reset network and context
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()
			qc := nw.GetFeeMarketClient()

			params := nw.App.FeeMarketKeeper.GetParams(ctx)
			exp := &types.QueryParamsResponse{Params: params}

			res, err := qc.Params(ctx.Context(), &types.QueryParamsRequest{})
			if tc.expPass {
				require.Equal(t, exp, res, tc.name)
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestQueryBaseFee(t *testing.T) {
	var (
		expRes         *types.QueryBaseFeeResponse
		nw             *network.UnitTestNetwork
		ctx            sdk.Context
		initialBaseFee sdkmath.LegacyDec
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"pass - default Base Fee",
			func() {
				expRes = &types.QueryBaseFeeResponse{BaseFee: &initialBaseFee}
			},
			true,
		},
		{
			"pass - non-nil Base Fee",
			func() {
				baseFee := sdkmath.LegacyNewDec(1)
				nw.App.FeeMarketKeeper.SetBaseFee(ctx, baseFee)

				expRes = &types.QueryBaseFeeResponse{BaseFee: &baseFee}
			},
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// reset network and context
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()
			qc := nw.GetFeeMarketClient()
			initialBaseFee = nw.App.FeeMarketKeeper.GetBaseFee(ctx)

			tc.malleate()

			res, err := qc.BaseFee(ctx.Context(), &types.QueryBaseFeeRequest{})
			if tc.expPass {
				require.NotNil(t, res)
				require.Equal(t, expRes, res, tc.name)
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestQueryBlockGas(t *testing.T) {
	var (
		nw  *network.UnitTestNetwork
		ctx sdk.Context
	)
	testCases := []struct {
		name    string
		expPass bool
	}{
		{
			"pass",
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// reset network and context
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()
			qc := nw.GetFeeMarketClient()

			gas := nw.App.FeeMarketKeeper.GetBlockGasWanted(ctx)
			exp := &types.QueryBlockGasResponse{Gas: int64(gas)} //#nosec G115

			res, err := qc.BlockGas(ctx.Context(), &types.QueryBlockGasRequest{})
			if tc.expPass {
				require.Equal(t, exp, res, tc.name)
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
