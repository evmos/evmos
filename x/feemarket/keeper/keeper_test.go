package keeper_test

import (
	_ "embed"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestSetGetBlockGasWanted() {
	testCases := []struct {
		name     string
		malleate func()
		expGas   uint64
	}{
		{
			"with last block given",
			func() {
				suite.app.FeeMarketKeeper.SetBlockGasWanted(suite.ctx, uint64(1000000))
			},
			uint64(1000000),
		},
	}
	for _, tc := range testCases {
		tc.malleate()

		gas := suite.app.FeeMarketKeeper.GetBlockGasWanted(suite.ctx)
		suite.Require().Equal(tc.expGas, gas, tc.name)
	}
}

func (suite *KeeperTestSuite) TestSetGetGasFee() {
	testCases := []struct {
		name     string
		malleate func()
		expFee   *big.Int
	}{
		{
			"with last block given",
			func() {
				suite.app.FeeMarketKeeper.SetBaseFee(suite.ctx, sdk.OneDec().BigInt())
			},
			sdk.OneDec().BigInt(),
		},
	}

	for _, tc := range testCases {
		tc.malleate()

		fee := suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx)
		suite.Require().Equal(tc.expFee, fee, tc.name)
	}
}
