package cosmos_test

import (
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	cosmosante "github.com/evmos/evmos/v17/app/ante/cosmos"
	"github.com/evmos/evmos/v17/testutil"
	testutiltx "github.com/evmos/evmos/v17/testutil/tx"
	"github.com/evmos/evmos/v17/utils"
)

var execTypes = []struct {
	name      string
	isCheckTx bool
	simulate  bool
}{
	{"deliverTx", false, false},
	{"deliverTxSimulate", false, true},
}

func (suite *AnteTestSuite) TestMinGasPriceDecorator() {
	denom := utils.BaseDenom
	testMsg := banktypes.MsgSend{
		FromAddress: "evmos1x8fhpj9nmhqk8z9kpgjt95ck2xwyue0ptzkucp",
		ToAddress:   "evmos1dx67l23hz9l0k9hcher8xz04uj7wf3yu26l2yn",
		Amount:      sdk.Coins{sdk.Coin{Amount: math.NewInt(10), Denom: denom}},
	}

	testCases := []struct {
		name                string
		malleate            func() sdk.Tx
		expPass             bool
		errMsg              string
		allowPassOnSimulate bool
	}{
		{
			"invalid cosmos tx type",
			func() sdk.Tx {
				return &testutiltx.InvalidTx{}
			},
			false,
			"invalid transaction type",
			false,
		},
		{
			"valid cosmos tx with MinGasPrices = 0, gasPrice = 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = math.LegacyZeroDec()
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				txBuilder := suite.CreateTestCosmosTxBuilder(math.NewInt(0), denom, &testMsg)
				return txBuilder.GetTx()
			},
			true,
			"",
			true,
		},
		{
			"valid cosmos tx with MinGasPrices = 0, gasPrice > 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = math.LegacyZeroDec()
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				txBuilder := suite.CreateTestCosmosTxBuilder(math.NewInt(10), denom, &testMsg)
				return txBuilder.GetTx()
			},
			true,
			"",
			true,
		},
		{
			"valid cosmos tx with MinGasPrices = 10, gasPrice = 10",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = math.LegacyNewDec(10)
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				txBuilder := suite.CreateTestCosmosTxBuilder(math.NewInt(10), denom, &testMsg)
				return txBuilder.GetTx()
			},
			true,
			"",
			true,
		},
		{
			"invalid cosmos tx with MinGasPrices = 10, gasPrice = 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = math.LegacyNewDec(10)
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				txBuilder := suite.CreateTestCosmosTxBuilder(math.NewInt(0), denom, &testMsg)
				return txBuilder.GetTx()
			},
			false,
			"provided fee < minimum global fee",
			true,
		},
		{
			"invalid cosmos tx with stake denom",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = math.LegacyNewDec(10)
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				txBuilder := suite.CreateTestCosmosTxBuilder(math.NewInt(10), sdk.DefaultBondDenom, &testMsg)
				return txBuilder.GetTx()
			},
			false,
			"provided fee < minimum global fee",
			true,
		},
		{
			"valid cosmos tx with MinGasPrices = 0, gasPrice = 0, valid fee",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = math.LegacyZeroDec()
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				txBuilder := suite.CreateTestCosmosTxBuilderWithFees(sdk.Coins{sdk.Coin{Amount: math.NewInt(0), Denom: denom}}, &testMsg)
				return txBuilder.GetTx()
			},
			true,
			"",
			true,
		},
		{
			"valid cosmos tx with MinGasPrices = 0, gasPrice = 0, nil fees, means len(fees) == 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = math.LegacyZeroDec()
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				txBuilder := suite.CreateTestCosmosTxBuilderWithFees(nil, &testMsg)
				return txBuilder.GetTx()
			},
			true,
			"",
			true,
		},
		{
			"valid cosmos tx with MinGasPrices = 0, gasPrice = 0, empty fees, means len(fees) == 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = math.LegacyZeroDec()
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				txBuilder := suite.CreateTestCosmosTxBuilderWithFees(sdk.Coins{}, &testMsg)
				return txBuilder.GetTx()
			},
			true,
			"",
			true,
		},
		{
			"valid cosmos tx with MinGasPrices = 0, gasPrice = 0, invalid fees",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = math.LegacyZeroDec()
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				fees := sdk.Coins{sdk.Coin{Amount: math.NewInt(0), Denom: denom}, sdk.Coin{Amount: math.NewInt(10), Denom: "stake"}}
				txBuilder := suite.CreateTestCosmosTxBuilderWithFees(fees, &testMsg)
				return txBuilder.GetTx()
			},
			false,
			fmt.Sprintf("expected only use native token %s for fee", denom),
			true,
		},
	}

	for _, et := range execTypes {
		for _, tc := range testCases {
			suite.Run(et.name+"_"+tc.name, func() {
				// s.SetupTest(et.isCheckTx)
				ctx := suite.ctx.WithIsReCheckTx(et.isCheckTx)
				dec := cosmosante.NewMinGasPriceDecorator(suite.app.FeeMarketKeeper, suite.app.EvmKeeper)
				_, err := dec.AnteHandle(ctx, tc.malleate(), et.simulate, testutil.NextFn)

				if (et.name == "deliverTx" && tc.expPass) || (et.name == "deliverTxSimulate" && et.simulate && tc.allowPassOnSimulate) {
					suite.Require().NoError(err, tc.name)
				} else {
					suite.Require().Error(err, tc.name)
					suite.Require().Contains(err.Error(), tc.errMsg, tc.name)
				}
			})
		}
	}
}
