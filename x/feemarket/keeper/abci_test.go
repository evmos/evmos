package keeper_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	"github.com/stretchr/testify/require"
)

func TestEndBlock(t *testing.T) {
	var (
		nw  *network.UnitTestNetwork
		ctx sdk.Context
	)

	testCases := []struct {
		name         string
		NoBaseFee    bool
		malleate     func()
		expGasWanted uint64
	}{
		{
			"baseFee nil",
			true,
			func() {},
			uint64(0),
		},
		{
			"pass",
			false,
			func() {
				meter := storetypes.NewGasMeter(uint64(1000000000))
				ctx = ctx.WithBlockGasMeter(meter)
				nw.App.FeeMarketKeeper.SetTransientBlockGasWanted(ctx, 5000000)
			},
			uint64(2500000),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// reset network and context
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()

			params := nw.App.FeeMarketKeeper.GetParams(ctx)
			params.NoBaseFee = tc.NoBaseFee

			err := nw.App.FeeMarketKeeper.SetParams(ctx, params)
			require.NoError(t, err)

			tc.malleate()

			err = nw.App.FeeMarketKeeper.EndBlock(ctx)
			require.NoError(t, err)

			gasWanted := nw.App.FeeMarketKeeper.GetBlockGasWanted(ctx)
			require.Equal(t, tc.expGasWanted, gasWanted, tc.name)
		})
	}
}
