package keeper_test

import (
	"math/big"
	"testing"

	"cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	"github.com/stretchr/testify/require"
)

func TestCalculateBaseFee(t *testing.T) {
	var (
		nw             *network.UnitTestNetwork
		ctx            sdk.Context
		initialBaseFee math.Int
	)

	testCases := []struct {
		name                 string
		NoBaseFee            bool
		blockHeight          int64
		parentBlockGasWanted uint64
		minGasPrice          math.LegacyDec
		expFee               func() *big.Int
	}{
		{
			"without BaseFee",
			true,
			0,
			0,
			math.LegacyZeroDec(),
			nil,
		},
		{
			"with BaseFee - initial EIP-1559 block",
			false,
			0,
			0,
			math.LegacyZeroDec(),
			func() *big.Int { return nw.App.FeeMarketKeeper.GetParams(ctx).BaseFee.BigInt() },
		},
		{
			"with BaseFee - parent block wanted the same gas as its target (ElasticityMultiplier = 2)",
			false,
			1,
			50,
			math.LegacyZeroDec(),
			func() *big.Int { return nw.App.FeeMarketKeeper.GetParams(ctx).BaseFee.BigInt() },
		},
		{
			"with BaseFee - parent block wanted the same gas as its target, with higher min gas price (ElasticityMultiplier = 2)",
			false,
			1,
			50,
			math.LegacyNewDec(1500000000),
			func() *big.Int { return nw.App.FeeMarketKeeper.GetParams(ctx).BaseFee.BigInt() },
		},
		{
			"with BaseFee - parent block wanted more gas than its target (ElasticityMultiplier = 2)",
			false,
			1,
			100,
			math.LegacyZeroDec(),
			func() *big.Int { return initialBaseFee.Add(math.NewInt(109375000)).BigInt() },
		},
		{
			"with BaseFee - parent block wanted more gas than its target, with higher min gas price (ElasticityMultiplier = 2)",
			false,
			1,
			100,
			math.LegacyNewDec(1500000000),
			func() *big.Int { return initialBaseFee.Add(math.NewInt(109375000)).BigInt() },
		},
		{
			"with BaseFee - Parent gas wanted smaller than parent gas target (ElasticityMultiplier = 2)",
			false,
			1,
			25,
			math.LegacyZeroDec(),
			func() *big.Int { return initialBaseFee.Sub(math.NewInt(54687500)).BigInt() },
		},
		{
			"with BaseFee - Parent gas wanted smaller than parent gas target, with higher min gas price (ElasticityMultiplier = 2)",
			false,
			1,
			25,
			math.LegacyNewDec(1500000000),
			func() *big.Int { return big.NewInt(1500000000) },
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// reset network and context
			nw = network.NewUnitTestNetwork()
			ctx = nw.GetContext()

			params := nw.App.FeeMarketKeeper.GetParams(ctx)
			params.NoBaseFee = tc.NoBaseFee
			params.MinGasPrice = tc.minGasPrice
			err := nw.App.FeeMarketKeeper.SetParams(ctx, params)
			require.NoError(t, err)

			initialBaseFee = params.BaseFee

			// Set block height
			ctx = ctx.WithBlockHeight(tc.blockHeight)

			// Set parent block gas
			nw.App.FeeMarketKeeper.SetBlockGasWanted(ctx, tc.parentBlockGasWanted)

			// Set next block target/gasLimit through Consensus Param MaxGas
			blockParams := tmproto.BlockParams{
				MaxGas:   100,
				MaxBytes: 10,
			}
			consParams := tmproto.ConsensusParams{Block: &blockParams}
			ctx = ctx.WithConsensusParams(consParams)

			fee := nw.App.FeeMarketKeeper.CalculateBaseFee(ctx)
			if tc.NoBaseFee {
				require.Nil(t, fee, tc.name)
			} else {
				require.Equal(t, tc.expFee(), fee, tc.name)
			}
		})
	}
}
