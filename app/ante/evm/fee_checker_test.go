package evm

import (
	"math/big"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/evmos/evmos/v19/encoding"
	"github.com/evmos/evmos/v19/types"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v19/x/feemarket/types"
)

var _ DynamicFeeEVMKeeper = MockEVMKeeper{}

type MockEVMKeeper struct {
	Params evmtypes.Params
}

func (m MockEVMKeeper) GetParams(_ sdk.Context) evmtypes.Params {
	return m.Params
}

func (m MockEVMKeeper) ChainID() *big.Int {
	return big.NewInt(9000)
}

type MockFeeMarketKeeper struct {
	BaseFee math.LegacyDec
}

func (m MockFeeMarketKeeper) GetBaseFee(_ sdk.Context) math.LegacyDec {
	return m.BaseFee
}

func (m MockFeeMarketKeeper) GetMinGasPrice(sdk.Context) (math.LegacyDec, error) {
	return math.LegacyZeroDec(), nil
}

func (m MockFeeMarketKeeper) AddTransientGasWanted(ctx sdk.Context, gasWanted uint64) (uint64, error) {
	return 0, nil
}

func (m MockFeeMarketKeeper) GetBaseFeeEnabled(ctx sdk.Context) bool {
	return true
}

func (m MockFeeMarketKeeper) GetParams(ctx sdk.Context) (params feemarkettypes.Params) {
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
	encodingConfig := encoding.MakeConfig(module.NewBasicManager())
	minGasPrices := sdk.NewDecCoins(sdk.NewDecCoin(evmtypes.DefaultEVMDenom, math.NewInt(10)))

	genesisCtx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	checkTxCtx := sdk.NewContext(nil, tmproto.Header{Height: 1}, true, log.NewNopLogger()).WithMinGasPrices(minGasPrices)
	deliverTxCtx := sdk.NewContext(nil, tmproto.Header{Height: 1}, false, log.NewNopLogger())

	defaultEvmParams := evmtypes.DefaultParams()
	notLondonParams := evmtypes.DefaultParams()
	londonBlockHeight := math.NewInt(100000)
	notLondonParams.ChainConfig.LondonBlock = &londonBlockHeight

	testCases := []struct {
		name        string
		ctx         sdk.Context
		evmKeeper   DynamicFeeEVMKeeper
		fmtkKeeper  FeeMarketKeeper
		buildTx     func() sdk.FeeTx
		expFees     string
		expPriority int64
		expSuccess  bool
	}{
		{
			"success, genesis tx",
			genesisCtx,
			MockEVMKeeper{Params: notLondonParams},
			MockFeeMarketKeeper{BaseFee: math.LegacyZeroDec()},
			func() sdk.FeeTx {
				return encodingConfig.TxConfig.NewTxBuilder().GetTx()
			},
			"",
			0,
			true,
		},
		{
			"fail, min-gas-prices",
			checkTxCtx,
			MockEVMKeeper{Params: notLondonParams},
			MockFeeMarketKeeper{BaseFee: math.LegacyZeroDec()},
			func() sdk.FeeTx {
				return encodingConfig.TxConfig.NewTxBuilder().GetTx()
			},
			"",
			0,
			false,
		},
		{
			"success, min-gas-prices",
			checkTxCtx,
			MockEVMKeeper{Params: notLondonParams},
			MockFeeMarketKeeper{BaseFee: math.LegacyZeroDec()},
			func() sdk.FeeTx {
				txBuilder := encodingConfig.TxConfig.NewTxBuilder()
				txBuilder.SetGasLimit(1)
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(10))))
				return txBuilder.GetTx()
			},
			"10aevmos",
			0,
			true,
		},
		{
			"success, min-gas-prices deliverTx",
			deliverTxCtx,
			MockEVMKeeper{Params: notLondonParams},
			MockFeeMarketKeeper{BaseFee: math.LegacyZeroDec()},
			func() sdk.FeeTx {
				return encodingConfig.TxConfig.NewTxBuilder().GetTx()
			},
			"",
			0,
			true,
		},
		{
			"fail, dynamic fee",
			deliverTxCtx,
			MockEVMKeeper{Params: defaultEvmParams},
			MockFeeMarketKeeper{BaseFee: math.LegacyOneDec()},
			func() sdk.FeeTx {
				txBuilder := encodingConfig.TxConfig.NewTxBuilder()
				txBuilder.SetGasLimit(1)
				return txBuilder.GetTx()
			},
			"",
			0,
			false,
		},
		{
			"success, dynamic fee",
			deliverTxCtx,
			MockEVMKeeper{Params: defaultEvmParams},
			MockFeeMarketKeeper{BaseFee: math.LegacyNewDec(10)},
			func() sdk.FeeTx {
				txBuilder := encodingConfig.TxConfig.NewTxBuilder()
				txBuilder.SetGasLimit(1)
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(10))))
				return txBuilder.GetTx()
			},
			"10aevmos",
			0,
			true,
		},
		{
			"success, dynamic fee priority",
			deliverTxCtx,
			MockEVMKeeper{Params: defaultEvmParams},
			MockFeeMarketKeeper{BaseFee: math.LegacyNewDec(10)},
			func() sdk.FeeTx {
				txBuilder := encodingConfig.TxConfig.NewTxBuilder()
				txBuilder.SetGasLimit(1)
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(10).Mul(evmtypes.DefaultPriorityReduction).Add(math.NewInt(10)))))
				return txBuilder.GetTx()
			},
			"10000010aevmos",
			10,
			true,
		},
		{
			"success, dynamic fee empty tipFeeCap",
			deliverTxCtx,
			MockEVMKeeper{Params: defaultEvmParams},
			MockFeeMarketKeeper{BaseFee: math.LegacyNewDec(10)},
			func() sdk.FeeTx {
				txBuilder := encodingConfig.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
				txBuilder.SetGasLimit(1)
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(10).Mul(evmtypes.DefaultPriorityReduction))))

				option, err := codectypes.NewAnyWithValue(&types.ExtensionOptionDynamicFeeTx{})
				require.NoError(t, err)
				txBuilder.SetExtensionOptions(option)
				return txBuilder.GetTx()
			},
			"10aevmos",
			0,
			true,
		},
		{
			"success, dynamic fee tipFeeCap",
			deliverTxCtx,
			MockEVMKeeper{Params: defaultEvmParams},
			MockFeeMarketKeeper{BaseFee: math.LegacyNewDec(10)},
			func() sdk.FeeTx {
				txBuilder := encodingConfig.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
				txBuilder.SetGasLimit(1)
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(10).Mul(evmtypes.DefaultPriorityReduction).Add(math.NewInt(10)))))

				option, err := codectypes.NewAnyWithValue(&types.ExtensionOptionDynamicFeeTx{
					MaxPriorityPrice: math.LegacyNewDec(5).MulInt(evmtypes.DefaultPriorityReduction),
				})
				require.NoError(t, err)
				txBuilder.SetExtensionOptions(option)
				return txBuilder.GetTx()
			},
			"5000010aevmos",
			5,
			true,
		},
		{
			"fail, negative dynamic fee tipFeeCap",
			deliverTxCtx,
			MockEVMKeeper{Params: defaultEvmParams},
			MockFeeMarketKeeper{BaseFee: math.LegacyNewDec(10)},
			func() sdk.FeeTx {
				txBuilder := encodingConfig.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
				txBuilder.SetGasLimit(1)
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(10).Mul(evmtypes.DefaultPriorityReduction).Add(math.NewInt(10)))))

				// set negative priority fee
				option, err := codectypes.NewAnyWithValue(&types.ExtensionOptionDynamicFeeTx{
					MaxPriorityPrice: math.LegacyNewDec(-5).MulInt(evmtypes.DefaultPriorityReduction),
				})
				require.NoError(t, err)
				txBuilder.SetExtensionOptions(option)
				return txBuilder.GetTx()
			},
			"",
			0,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fees, priority, err := NewDynamicFeeChecker(tc.evmKeeper, tc.fmtkKeeper)(tc.ctx, tc.buildTx())
			if tc.expSuccess {
				require.Equal(t, tc.expFees, fees.String())
				require.Equal(t, tc.expPriority, priority)
			} else {
				require.Error(t, err)
			}
		})
	}
}
