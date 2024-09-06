package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v19/x/feemarket/types"
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
		aux            sdkmath.Int
		expRes         *types.QueryBaseFeeResponse
		nw             *network.UnitTestNetwork
		ctx            sdk.Context
		initialBaseFee sdkmath.Int
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
				baseFee := sdkmath.OneInt().BigInt()
				nw.App.FeeMarketKeeper.SetBaseFee(ctx, baseFee)

				aux = sdkmath.NewIntFromBigInt(baseFee)
				expRes = &types.QueryBaseFeeResponse{BaseFee: &aux}
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
			initialBaseFee = sdkmath.NewIntFromBigInt(nw.App.FeeMarketKeeper.GetBaseFee(ctx))

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
