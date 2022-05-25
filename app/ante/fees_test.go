package ante_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/v4/app/ante"
)

var execTypes = []struct {
	name      string
	isCheckTx bool
	simulate  bool
}{
	{"deliverTx", false, false},
	{"deliverTxSimulate", false, true},
}

func nextFn(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
	return ctx, nil
}

func (s AnteTestSuite) TestMinGasPriceDecorator() {
	testMsg := banktypes.MsgSend{
		FromAddress: "evmos1x8fhpj9nmhqk8z9kpgjt95ck2xwyue0ptzkucp",
		ToAddress:   "evmos1dx67l23hz9l0k9hcher8xz04uj7wf3yu26l2yn",
		Amount:      sdk.Coins{sdk.Coin{Amount: sdk.NewInt(10), Denom: s.denom}},
	}

	testCases := []struct {
		name     string
		malleate func() sdk.Tx
		expPass  bool
		errMsg   string
	}{
		{
			"invalid cosmos tx type",
			func() sdk.Tx {
				return &invalidTx{}
			},
			false,
			"must be a FeeTx",
		},
		{
			"valid cosmos tx with MinGasPrices = 0, gasPrice = 0",
			func() sdk.Tx {
				params := s.app.FeesKeeper.GetParams(s.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				s.app.FeesKeeper.SetParams(s.ctx, params)

				txBuilder := s.CreateTestTxBuilder(sdk.NewInt(0), s.denom, &testMsg)
				return txBuilder.GetTx()
			},
			true,
			"",
		},
		{
			"valid cosmos tx with MinGasPrices = 10, gasPrice = 10",
			func() sdk.Tx {
				params := s.app.FeesKeeper.GetParams(s.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				s.app.FeesKeeper.SetParams(s.ctx, params)

				txBuilder := s.CreateTestTxBuilder(sdk.NewInt(10), s.denom, &testMsg)
				return txBuilder.GetTx()
			},
			true,
			"",
		},
		{
			"invalid cosmos tx with MinGasPrices = 10, gasPrice = 0",
			func() sdk.Tx {
				params := s.app.FeesKeeper.GetParams(s.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				s.app.FeesKeeper.SetParams(s.ctx, params)

				txBuilder := s.CreateTestTxBuilder(sdk.NewInt(0), s.denom, &testMsg)
				return txBuilder.GetTx()
			},
			false,
			"provided fee < minimum global fee",
		},
		{
			"invalid cosmos tx with wrong denom",
			func() sdk.Tx {
				params := s.app.FeesKeeper.GetParams(s.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				s.app.FeesKeeper.SetParams(s.ctx, params)

				txBuilder := s.CreateTestTxBuilder(sdk.NewInt(10), "stake", &testMsg)
				return txBuilder.GetTx()
			},
			false,
			"provided fee < minimum global fee",
		},
	}

	for _, et := range execTypes {
		for _, tc := range testCases {
			s.Run(et.name+"_"+tc.name, func() {
				s.SetupTest(et.isCheckTx)
				dec := ante.NewMinGasPriceDecorator(s.app.FeesKeeper, s.app.EvmKeeper)
				_, err := dec.AnteHandle(s.ctx, tc.malleate(), et.simulate, nextFn)

				if tc.expPass {
					s.Require().NoError(err, tc.name)
				} else {
					s.Require().Error(err, tc.name)
					s.Require().Contains(err.Error(), tc.errMsg, tc.name)
				}
			})
		}
	}
}

func (s AnteTestSuite) TestEthMinGasPriceDecorator() {
	from := tests.GenerateAddress()
	to := tests.GenerateAddress()
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
				params := s.app.FeesKeeper.GetParams(s.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				s.app.FeesKeeper.SetParams(s.ctx, params)
				return &invalidTx{}
			},
			false,
			"invalid message type",
		},
		{
			"wrong tx type",
			func() sdk.Tx {
				params := s.app.FeesKeeper.GetParams(s.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				s.app.FeesKeeper.SetParams(s.ctx, params)
				testMsg := banktypes.MsgSend{
					FromAddress: "evmos1x8fhpj9nmhqk8z9kpgjt95ck2xwyue0ptzkucp",
					ToAddress:   "evmos1dx67l23hz9l0k9hcher8xz04uj7wf3yu26l2yn",
					Amount:      sdk.Coins{sdk.Coin{Amount: sdk.NewInt(10), Denom: s.denom}},
				}
				txBuilder := s.CreateTestTxBuilder(sdk.NewInt(0), s.denom, &testMsg)
				return txBuilder.GetTx()
			},
			false,
			"invalid message type",
		},
		{
			"valid: invalid tx type with MinGasPrices = 0",
			func() sdk.Tx {
				params := s.app.FeesKeeper.GetParams(s.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				s.app.FeesKeeper.SetParams(s.ctx, params)
				return &invalidTx{}
			},
			true,
			"",
		},
		{
			"valid legacy tx with MinGasPrices = 0, gasPrice = 0",
			func() sdk.Tx {
				params := s.app.FeesKeeper.GetParams(s.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				s.app.FeesKeeper.SetParams(s.ctx, params)

				msg := s.BuildTestEthTx(from, to, big.NewInt(0), nil, nil, nil)
				txBuilder := s.CreateEthTestTxBuilder(msg)
				return txBuilder.GetTx()
			},
			true,
			"",
		},
		{
			"valid legacy tx with MinGasPrices = 10, gasPrice = 10",
			func() sdk.Tx {
				params := s.app.FeesKeeper.GetParams(s.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				s.app.FeesKeeper.SetParams(s.ctx, params)

				msg := s.BuildTestEthTx(from, to, big.NewInt(10), nil, nil, nil)
				txBuilder := s.CreateEthTestTxBuilder(msg)
				return txBuilder.GetTx()
			},
			true,
			"",
		},
		{
			"invalid legacy tx with MinGasPrices = 10, gasPrice = 0",
			func() sdk.Tx {
				params := s.app.FeesKeeper.GetParams(s.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				s.app.FeesKeeper.SetParams(s.ctx, params)

				msg := s.BuildTestEthTx(from, to, big.NewInt(0), nil, nil, nil)
				txBuilder := s.CreateEthTestTxBuilder(msg)
				return txBuilder.GetTx()
			},
			false,
			"provided fee < minimum global fee",
		},
		{
			"valid dynamic tx with MinGasPrices = 0, EffectivePrice = 0",
			func() sdk.Tx {
				params := s.app.FeesKeeper.GetParams(s.ctx)
				params.MinGasPrice = sdk.ZeroDec()
				s.app.FeesKeeper.SetParams(s.ctx, params)

				msg := s.BuildTestEthTx(from, to, nil, big.NewInt(0), big.NewInt(0), &emptyAccessList)
				txBuilder := s.CreateEthTestTxBuilder(msg)
				return txBuilder.GetTx()
			},
			true,
			"",
		},
		{
			"valid dynamic tx with MinGasPrices < EffectivePrice",
			func() sdk.Tx {
				params := s.app.FeesKeeper.GetParams(s.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				s.app.FeesKeeper.SetParams(s.ctx, params)

				msg := s.BuildTestEthTx(from, to, nil, big.NewInt(100), big.NewInt(100), &emptyAccessList)
				txBuilder := s.CreateEthTestTxBuilder(msg)
				return txBuilder.GetTx()
			},
			true,
			"",
		},
		{
			"invalid dynamic tx with MinGasPrices > EffectivePrice",
			func() sdk.Tx {
				params := s.app.FeesKeeper.GetParams(s.ctx)
				params.MinGasPrice = sdk.NewDec(10)
				s.app.FeesKeeper.SetParams(s.ctx, params)

				msg := s.BuildTestEthTx(from, to, nil, big.NewInt(0), big.NewInt(0), &emptyAccessList)
				txBuilder := s.CreateEthTestTxBuilder(msg)
				return txBuilder.GetTx()
			},
			false,
			"provided fee < minimum global fee",
		},
		{
			"invalid dynamic tx with MinGasPrices > BaseFee, MinGasPrices > EffectivePrice",
			func() sdk.Tx {
				params := s.app.FeesKeeper.GetParams(s.ctx)
				params.MinGasPrice = sdk.NewDec(100)
				s.app.FeesKeeper.SetParams(s.ctx, params)

				feemarketParams := s.app.FeeMarketKeeper.GetParams(s.ctx)
				feemarketParams.BaseFee = sdk.NewInt(10)
				s.app.FeeMarketKeeper.SetParams(s.ctx, feemarketParams)

				msg := s.BuildTestEthTx(from, to, nil, big.NewInt(1000), big.NewInt(0), &emptyAccessList)
				txBuilder := s.CreateEthTestTxBuilder(msg)
				return txBuilder.GetTx()
			},
			false,
			"provided fee < minimum global fee",
		},
		{
			"valid dynamic tx with MinGasPrices > BaseFee, MinGasPrices < EffectivePrice (big GasTipCap)",
			func() sdk.Tx {
				params := s.app.FeesKeeper.GetParams(s.ctx)
				params.MinGasPrice = sdk.NewDec(100)
				s.app.FeesKeeper.SetParams(s.ctx, params)

				feemarketParams := s.app.FeeMarketKeeper.GetParams(s.ctx)
				feemarketParams.BaseFee = sdk.NewInt(10)
				s.app.FeeMarketKeeper.SetParams(s.ctx, feemarketParams)

				msg := s.BuildTestEthTx(from, to, nil, big.NewInt(1000), big.NewInt(101), &emptyAccessList)
				txBuilder := s.CreateEthTestTxBuilder(msg)
				return txBuilder.GetTx()
			},
			true,
			"",
		},
	}

	for _, et := range execTypes {
		for _, tc := range testCases {
			s.Run(et.name+"_"+tc.name, func() {
				s.SetupTest(et.isCheckTx)
				dec := ante.NewEthMinGasPriceDecorator(s.app.FeesKeeper, s.app.EvmKeeper)
				_, err := dec.AnteHandle(s.ctx, tc.malleate(), et.simulate, nextFn)

				if tc.expPass {
					s.Require().NoError(err, tc.name)
				} else {
					s.Require().Error(err, tc.name)
					s.Require().Contains(err.Error(), tc.errMsg, tc.name)
				}
			})
		}
	}
}
