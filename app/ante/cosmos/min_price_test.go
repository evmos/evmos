package cosmos_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	cosmosante "github.com/evmos/evmos/v15/app/ante/cosmos"
	"github.com/evmos/evmos/v15/cmd/config"
	"github.com/evmos/evmos/v15/testutil"
	testutiltx "github.com/evmos/evmos/v15/testutil/tx"
	"github.com/evmos/evmos/v15/utils"
)

var execTypes = []struct {
	name      string
	isCheckTx bool
	simulate  bool
}{
	{"checkTx", true, false},
	{"deliverTx", false, false},
	{"deliverTxSimulate", false, true},
}

func (suite *AnteTestSuite) TestMinGasPriceDecorator() {
	denom := utils.BaseDenom
	testMsg := banktypes.MsgSend{
		FromAddress: "evmos1x8fhpj9nmhqk8z9kpgjt95ck2xwyue0ptzkucp",
		ToAddress:   "evmos1dx67l23hz9l0k9hcher8xz04uj7wf3yu26l2yn",
		Amount:      sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(10), Denom: denom}},
	}

	testCases := []struct {
		name                string
		malleate            func() sdk.Tx
		expPass             bool
		errMsg              string
		localMinGasPrice    int64
		expCheckPass        bool
		allowPassOnSimulate bool
	}{
		{
			"invalid cosmos tx type",
			func() sdk.Tx {
				return &testutiltx.InvalidTx{}
			},
			false,
			"invalid transaction type",
			0,
			false,
			false,
		},
		{
			"valid cosmos tx with MinGasPrices = 0, gasPrice = 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(0), denom, &testMsg)
				return txBuilder.GetTx()
			},
			true,
			"",
			0,
			true,
			true,
		},
		{
			"valid cosmos tx with MinGasPrices = 0, gasPrice > 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(10), denom, &testMsg)
				return txBuilder.GetTx()
			},
			true,
			"",
			0,
			true,
			true,
		},
		{
			"valid cosmos tx with MinGasPrices = 10, gasPrice = 10",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(10), denom, &testMsg)
				return txBuilder.GetTx()
			},
			true,
			"",
			0,
			true,
			true,
		},
		{
			"invalid cosmos tx with MinGasPrices = 10, gasPrice = 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(0), denom, &testMsg)
				return txBuilder.GetTx()
			},
			false,
			"provided fee < minimum global fee",
			0,
			false,
			true,
		},
		{
			"invalid cosmos tx with stake denom",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(10), sdk.DefaultBondDenom, &testMsg)
				return txBuilder.GetTx()
			},
			false,
			"provided fee < minimum global fee",
			0,
			false,
			true,
		},
		{
			"valid cosmos tx with MinGasPrices = 0, gasPrice = 0, valid fee",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				txBuilder := suite.CreateTestCosmosTxBuilderWithFees(sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(0), Denom: denom}}, &testMsg)
				return txBuilder.GetTx()
			},
			true,
			"",
			0,
			true,
			true,
		},
		{
			"valid cosmos tx with MinGasPrices = 0, gasPrice = 0, nil fees, means len(fees) == 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				txBuilder := suite.CreateTestCosmosTxBuilderWithFees(nil, &testMsg)
				return txBuilder.GetTx()
			},
			true,
			"",
			0,
			true,
			true,
		},
		{
			"valid cosmos tx with MinGasPrices = 0, gasPrice = 0, empty fees, means len(fees) == 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				txBuilder := suite.CreateTestCosmosTxBuilderWithFees(sdk.Coins{}, &testMsg)
				return txBuilder.GetTx()
			},
			true,
			"provided fee < minimum global fee",
			0,
			true,
			true,
		},
		{
			"valid cosmos tx with MinGasPrices = 0, gasPrice = 0, invalid fees",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				fees := sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(0), Denom: denom}, sdk.Coin{Amount: sdkmath.NewInt(10), Denom: "stake"}}
				txBuilder := suite.CreateTestCosmosTxBuilderWithFees(fees, &testMsg)
				return txBuilder.GetTx()
			},
			false,
			fmt.Sprintf("expected only use native token %s for fee", denom),
			0,
			false,
			true,
		},
		{
			"valid cosmos tx with MinGasPrices = 10, LocalMinGasPrices = 20, gasPrice = 10",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(10), denom, &testMsg)
				return txBuilder.GetTx()
			},
			true,
			"provided fee < minimum local fee",
			20,
			false,
			true,
		},
	}

	for _, et := range execTypes {
		for _, tc := range testCases {
			suite.Run(et.name+"_"+tc.name, func() {
				// s.SetupTest(et.isCheckTx)
				ctx := suite.ctx.WithIsReCheckTx(et.isCheckTx)

				localMinGasPrices := sdk.NewDecCoins(sdk.NewDecCoinFromDec(config.BaseDenom, sdk.NewDec(tc.localMinGasPrice)))
				ctx = ctx.WithMinGasPrices(localMinGasPrices)

				dec := cosmosante.NewMinGasPriceDecorator(suite.app.FeeMarketKeeper, suite.app.EvmKeeper)
				_, err := dec.AnteHandle(ctx, tc.malleate(), et.simulate, testutil.NextFn)

				if (et.name == "checkTx" && tc.expCheckPass) ||
					(et.name == "deliverTx" && tc.expPass) ||
					(et.name == "deliverTxSimulate" && et.simulate && tc.allowPassOnSimulate) {
					suite.Require().NoError(err, tc.name)
				} else {
					suite.Require().Error(err, tc.name)
					suite.Require().Contains(err.Error(), tc.errMsg, tc.name)
				}
			})
		}
	}
}
