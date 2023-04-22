package evm_test

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmante "github.com/evmos/evmos/v13/app/ante/evm"
	"github.com/evmos/evmos/v13/testutil"
	testutiltx "github.com/evmos/evmos/v13/testutil/tx"
	evmtypes "github.com/evmos/evmos/v13/x/evm/types"
)

var execTypes = []struct {
	name      string
	isCheckTx bool
	simulate  bool
}{
	{"deliverTx", false, false},
	{"deliverTxSimulate", false, true},
}

func (suite *AnteTestSuite) TestEthMinGasPriceDecorator() {
	denom := evmtypes.DefaultEVMDenom
	from, privKey := testutiltx.NewAddrKey()
	to := testutiltx.GenerateAddress()
	emptyAccessList := ethtypes.AccessList{}

	testCases := []struct {
		name     string
		malleate func() sdk.Tx
		expPass  bool
		errMsg   string
	}{
		{
			"invalid tx type",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				return &testutiltx.InvalidTx{}
			},
			false,
			"invalid message type",
		},
		{
			"wrong tx type",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				testMsg := banktypes.MsgSend{
					FromAddress: "evmos1x8fhpj9nmhqk8z9kpgjt95ck2xwyue0ptzkucp",
					ToAddress:   "evmos1dx67l23hz9l0k9hcher8xz04uj7wf3yu26l2yn",
					Amount:      sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(10), Denom: denom}},
				}
				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(0), denom, &testMsg)
				return txBuilder.GetTx()
			},
			false,
			"invalid message type",
		},
		{
			"valid: invalid tx type with MinGasPrices = 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				return &testutiltx.InvalidTx{}
			},
			true,
			"",
		},
		{
			"valid legacy tx with MinGasPrices = 0, gasPrice = 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), big.NewInt(0), nil, nil, nil)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			true,
			"",
		},
		{
			"valid legacy tx with MinGasPrices = 0, gasPrice > 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), big.NewInt(10), nil, nil, nil)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			true,
			"",
		},
		{
			"valid legacy tx with MinGasPrices = 10, gasPrice = 10",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), big.NewInt(10), nil, nil, nil)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			true,
			"",
		},
		{
			"invalid legacy tx with MinGasPrices = 10, gasPrice = 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), big.NewInt(0), nil, nil, nil)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			false,
			"provided fee < minimum global fee",
		},
		{
			"valid dynamic tx with MinGasPrices = 0, EffectivePrice = 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), nil, big.NewInt(0), big.NewInt(0), &emptyAccessList)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			true,
			"",
		},
		{
			"valid dynamic tx with MinGasPrices = 0, EffectivePrice > 0",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), nil, big.NewInt(100), big.NewInt(50), &emptyAccessList)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			true,
			"",
		},
		{
			"valid dynamic tx with MinGasPrices < EffectivePrice",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), nil, big.NewInt(100), big.NewInt(100), &emptyAccessList)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			true,
			"",
		},
		{
			"invalid dynamic tx with MinGasPrices > EffectivePrice",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), nil, big.NewInt(0), big.NewInt(0), &emptyAccessList)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			false,
			"provided fee < minimum global fee",
		},
		{
			"invalid dynamic tx with MinGasPrices > BaseFee, MinGasPrices > EffectivePrice",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(100)
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				feemarketParams := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				feemarketParams.BaseFee = sdkmath.NewInt(10)
				err = suite.app.FeeMarketKeeper.SetParams(suite.ctx, feemarketParams)
				suite.Require().NoError(err)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), nil, big.NewInt(1000), big.NewInt(0), &emptyAccessList)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			false,
			"provided fee < minimum global fee",
		},
		{
			"valid dynamic tx with MinGasPrices > BaseFee, MinGasPrices < EffectivePrice (big GasTipCap)",
			func() sdk.Tx {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.MinGasPrice = sdk.NewDec(100)
				err := suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)

				feemarketParams := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				feemarketParams.BaseFee = sdkmath.NewInt(10)
				err = suite.app.FeeMarketKeeper.SetParams(suite.ctx, feemarketParams)
				suite.Require().NoError(err)

				msg := suite.BuildTestEthTx(from, to, nil, make([]byte, 0), nil, big.NewInt(1000), big.NewInt(101), &emptyAccessList)
				return suite.CreateTestTx(msg, privKey, 1, false)
			},
			true,
			"",
		},
	}

	for _, et := range execTypes {
		for _, tc := range testCases {
			suite.Run(et.name+"_"+tc.name, func() {
				// s.SetupTest(et.isCheckTx)
				suite.SetupTest()
				dec := evmante.NewEthMinGasPriceDecorator(suite.app.FeeMarketKeeper, suite.app.EvmKeeper)
				_, err := dec.AnteHandle(suite.ctx, tc.malleate(), et.simulate, testutil.NextFn)

				if tc.expPass {
					suite.Require().NoError(err, tc.name)
				} else {
					suite.Require().Error(err, tc.name)
					suite.Require().Contains(err.Error(), tc.errMsg, tc.name)
				}
			})
		}
	}
}

func (suite *AnteTestSuite) TestEthMempoolFeeDecorator() {
	// TODO: add test
}
