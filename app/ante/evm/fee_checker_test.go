package evm_test

import (
	"math/big"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/evmos/evmos/v20/app/ante/evm"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	"github.com/evmos/evmos/v20/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v20/x/feemarket/types"
)

var _ evm.FeeMarketKeeper = MockFeemarketKeeper{}

type MockFeemarketKeeper struct {
	BaseFee math.LegacyDec
}

func (m MockFeemarketKeeper) GetBaseFee(_ sdk.Context) math.LegacyDec {
	return m.BaseFee
}

func (m MockFeemarketKeeper) GetBaseFeeEnabled(_ sdk.Context) bool {
	return true
}

func (m MockFeemarketKeeper) AddTransientGasWanted(_ sdk.Context, _ uint64) (uint64, error) {
	return 0, nil
}

func (m MockFeemarketKeeper) GetParams(_ sdk.Context) (params feemarkettypes.Params) {
	return feemarkettypes.DefaultParams()
}

func TestSDKTxFeeChecker(t *testing.T) {
	// testCases:
	//   fallback
	//      genesis tx
	//      checkTx, validate with min-gas-prices
	//      deliverTx, no validation
	//   dynamic fee
	//      with extension option
	//      without extension option
	//      london hardfork enableness
	nw := network.New()
	encodingConfig := nw.GetEncodingConfig()
	baseDenom := evmtypes.GetEVMCoinDenom()
	minGasPrices := sdk.NewDecCoins(sdk.NewDecCoin(baseDenom, math.NewInt(10)))

	genesisCtx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	checkTxCtx := sdk.NewContext(nil, tmproto.Header{Height: 1}, true, log.NewNopLogger()).WithMinGasPrices(minGasPrices)
	deliverTxCtx := sdk.NewContext(nil, tmproto.Header{Height: 1}, false, log.NewNopLogger())

	testCases := []struct {
		name          string
		ctx           sdk.Context
		keeper        evm.FeeMarketKeeper
		buildTx       func() sdk.FeeTx
		londonEnabled bool
		expFees       string
		expPriority   int64
		expSuccess    bool
	}{
		{
			"success, genesis tx",
			genesisCtx,
			MockFeemarketKeeper{},
			func() sdk.FeeTx {
				return encodingConfig.TxConfig.NewTxBuilder().GetTx()
			},
			false,
			"",
			0,
			true,
		},
		{
			"fail, min-gas-prices",
			checkTxCtx,
			MockFeemarketKeeper{},
			func() sdk.FeeTx {
				return encodingConfig.TxConfig.NewTxBuilder().GetTx()
			},
			false,
			"",
			0,
			false,
		},
		{
			"success, min-gas-prices",
			checkTxCtx,
			MockFeemarketKeeper{},
			func() sdk.FeeTx {
				txBuilder := encodingConfig.TxConfig.NewTxBuilder()
				txBuilder.SetGasLimit(1)
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(baseDenom, math.NewInt(10))))
				return txBuilder.GetTx()
			},
			false,
			"10aevmos",
			0,
			true,
		},
		{
			"success, min-gas-prices deliverTx",
			deliverTxCtx,
			MockFeemarketKeeper{},
			func() sdk.FeeTx {
				return encodingConfig.TxConfig.NewTxBuilder().GetTx()
			},
			false,
			"",
			0,
			true,
		},
		{
			"fail, dynamic fee",
			deliverTxCtx,
			MockFeemarketKeeper{
				BaseFee: math.LegacyNewDec(1),
			},
			func() sdk.FeeTx {
				txBuilder := encodingConfig.TxConfig.NewTxBuilder()
				txBuilder.SetGasLimit(1)
				return txBuilder.GetTx()
			},
			true,
			"",
			0,
			false,
		},
		{
			"success, dynamic fee",
			deliverTxCtx,
			MockFeemarketKeeper{
				BaseFee: math.LegacyNewDec(10),
			},
			func() sdk.FeeTx {
				txBuilder := encodingConfig.TxConfig.NewTxBuilder()
				txBuilder.SetGasLimit(1)
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(baseDenom, math.NewInt(10))))
				return txBuilder.GetTx()
			},
			true,
			"10aevmos",
			0,
			true,
		},
		{
			"success, dynamic fee priority",
			deliverTxCtx,
			MockFeemarketKeeper{
				BaseFee: math.LegacyNewDec(10),
			},
			func() sdk.FeeTx {
				txBuilder := encodingConfig.TxConfig.NewTxBuilder()
				txBuilder.SetGasLimit(1)
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(baseDenom, math.NewInt(10).Mul(evmtypes.DefaultPriorityReduction).Add(math.NewInt(10)))))
				return txBuilder.GetTx()
			},
			true,
			"10000010aevmos",
			10,
			true,
		},
		{
			"success, dynamic fee empty tipFeeCap",
			deliverTxCtx,
			MockFeemarketKeeper{
				BaseFee: math.LegacyNewDec(10),
			},
			func() sdk.FeeTx {
				txBuilder := encodingConfig.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
				txBuilder.SetGasLimit(1)
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(baseDenom, math.NewInt(10).Mul(evmtypes.DefaultPriorityReduction))))

				option, err := codectypes.NewAnyWithValue(&types.ExtensionOptionDynamicFeeTx{})
				require.NoError(t, err)
				txBuilder.SetExtensionOptions(option)
				return txBuilder.GetTx()
			},
			true,
			"10aevmos",
			0,
			true,
		},
		{
			"success, dynamic fee tipFeeCap",
			deliverTxCtx,
			MockFeemarketKeeper{
				BaseFee: math.LegacyNewDec(10),
			},
			func() sdk.FeeTx {
				txBuilder := encodingConfig.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
				txBuilder.SetGasLimit(1)
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(baseDenom, math.NewInt(10).Mul(evmtypes.DefaultPriorityReduction).Add(math.NewInt(10)))))

				option, err := codectypes.NewAnyWithValue(&types.ExtensionOptionDynamicFeeTx{
					MaxPriorityPrice: math.LegacyNewDec(5).MulInt(evmtypes.DefaultPriorityReduction),
				})
				require.NoError(t, err)
				txBuilder.SetExtensionOptions(option)
				return txBuilder.GetTx()
			},
			true,
			"5000010aevmos",
			5,
			true,
		},
		{
			"fail, negative dynamic fee tipFeeCap",
			deliverTxCtx,
			MockFeemarketKeeper{
				BaseFee: math.LegacyNewDec(10),
			},
			func() sdk.FeeTx {
				txBuilder := encodingConfig.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
				txBuilder.SetGasLimit(1)
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(baseDenom, math.NewInt(10).Mul(evmtypes.DefaultPriorityReduction).Add(math.NewInt(10)))))

				// set negative priority fee
				option, err := codectypes.NewAnyWithValue(&types.ExtensionOptionDynamicFeeTx{
					MaxPriorityPrice: math.LegacyNewDec(-5).MulInt(evmtypes.DefaultPriorityReduction),
				})
				require.NoError(t, err)
				txBuilder.SetExtensionOptions(option)
				return txBuilder.GetTx()
			},
			true,
			"",
			0,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := evmtypes.GetChainConfig()
			if !tc.londonEnabled {
				cfg.LondonBlock = big.NewInt(10000)
			} else {
				cfg.LondonBlock = big.NewInt(0)
			}
			fees, priority, err := evm.NewDynamicFeeChecker(tc.keeper)(tc.ctx, tc.buildTx())
			if tc.expSuccess {
				require.Equal(t, tc.expFees, fees.String())
				require.Equal(t, tc.expPriority, priority)
			} else {
				require.Error(t, err)
			}
		})
	}
}
