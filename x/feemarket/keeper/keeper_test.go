package keeper_test

import (
	"math/big"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	"github.com/stretchr/testify/require"
)

func TestSetGetBlockGasWanted(t *testing.T) {
	var (
		nw  *network.UnitTestNetwork
		ctx sdk.Context
	)
	testCases := []struct {
		name     string
		malleate func()
		expGas   uint64
	}{
		{
			"with last block given",
			func() {
				nw.App.FeeMarketKeeper.SetBlockGasWanted(ctx, uint64(1000000))
			},
			uint64(1000000),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// reset network and context
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()

			tc.malleate()

			gas := nw.App.FeeMarketKeeper.GetBlockGasWanted(ctx)
			require.Equal(t, tc.expGas, gas, tc.name)
		})
	}
}

func TestSetGetGasFee(t *testing.T) {
	var (
		nw  *network.UnitTestNetwork
		ctx sdk.Context
	)
	testCases := []struct {
		name     string
		malleate func()
		expFee   *big.Int
	}{
		{
			"with last block given",
			func() {
				nw.App.FeeMarketKeeper.SetBaseFee(ctx, math.LegacyOneDec().BigInt())
			},
			math.LegacyOneDec().BigInt(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// reset network and context
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()

			tc.malleate()

			fee := nw.App.FeeMarketKeeper.GetBaseFee(ctx)
			require.Equal(t, tc.expFee, fee, tc.name)
		})
	}
}
