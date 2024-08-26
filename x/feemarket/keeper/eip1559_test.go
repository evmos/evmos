package keeper_test

import (
	"fmt"

	"cosmossdk.io/math"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

func (suite *KeeperTestSuite) TestCalculateBaseFee() {
	defaultBaseFee := suite.app.FeeMarketKeeper.GetParams(suite.ctx).BaseFee
	baseFee1 := math.LegacyNewDecWithPrec(110, 2)
	baseFee3 := math.LegacyNewDecWithPrec(9375, 5)
	baseFee4 := math.LegacyNewDecWithPrec(150000000000, 2)

	testCases := []struct {
		name                 string
		NoBaseFee            bool
		blockHeight          int64
		parentBlockGasWanted uint64
		minGasPrice          math.LegacyDec
		expFee               *math.LegacyDec
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
			&defaultBaseFee,
		},
		{
			"with BaseFee - parent block wanted the same gas as its target (ElasticityMultiplier = 2)",
			false,
			1,
			50,
			math.LegacyZeroDec(),
			&defaultBaseFee,
		},
		{
			"with BaseFee - parent block wanted the same gas as its target, with higher min gas price (ElasticityMultiplier = 2)",
			false,
			1,
			50,
			math.LegacyNewDec(1500000000),
			&defaultBaseFee,
		},
		{
			"with BaseFee - parent block wanted more gas than its target (ElasticityMultiplier = 2)",
			false,
			1,
			100,
			math.LegacyZeroDec(),
			&baseFee1,
		},
		{
			"with BaseFee - parent block wanted more gas than its target, with higher min gas price (ElasticityMultiplier = 2)",
			false,
			1,
			100,
			math.LegacyNewDec(1500000000),
			&baseFee1,
		},
		{
			"with BaseFee - Parent gas wanted smaller than parent gas target (ElasticityMultiplier = 2)",
			false,
			1,
			25,
			math.LegacyZeroDec(),
			&baseFee3,
		},
		{
			"with BaseFee - Parent gas wanted smaller than parent gas target, with higher min gas price (ElasticityMultiplier = 2)",
			false,
			1,
			25,
			math.LegacyNewDec(1500000000),
			&baseFee4,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
			params.NoBaseFee = tc.NoBaseFee
			params.MinGasPrice = tc.minGasPrice

			err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
			suite.Require().NoError(err)

			// Set block height
			suite.ctx = suite.ctx.WithBlockHeight(tc.blockHeight)

			// Set parent block gas
			suite.app.FeeMarketKeeper.SetBlockGasWanted(suite.ctx, tc.parentBlockGasWanted)

			// Set next block target/gasLimit through Consensus Param MaxGas
			blockParams := tmproto.BlockParams{
				MaxGas:   100,
				MaxBytes: 10,
			}
			consParams := tmproto.ConsensusParams{Block: &blockParams}
			suite.ctx = suite.ctx.WithConsensusParams(&consParams)

			fee := suite.app.FeeMarketKeeper.CalculateBaseFee(suite.ctx)
			if tc.NoBaseFee {
				suite.Require().Equal(math.LegacyDec{}, fee, tc.name)
			} else {
				suite.Require().Equal(tc.expFee, &fee, tc.name)
			}
		})
	}
}
