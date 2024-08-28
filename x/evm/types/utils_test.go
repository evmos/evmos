package types_test

import (
	"errors"
	"math/big"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	"github.com/ethereum/go-ethereum/common"

	proto "github.com/cosmos/gogoproto/proto"

	"github.com/evmos/evmos/v19/app"
	"github.com/evmos/evmos/v19/encoding"
	utiltx "github.com/evmos/evmos/v19/testutil/tx"
	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/evm/types"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"

	"github.com/stretchr/testify/require"
)

func TestEvmDataEncoding(t *testing.T) {
	ret := []byte{0x5, 0x8}

	data := &evmtypes.MsgEthereumTxResponse{
		Hash: common.BytesToHash([]byte("hash")).String(),
		Logs: []*evmtypes.Log{{
			Data:        []byte{1, 2, 3, 4},
			BlockNumber: 17,
		}},
		Ret: ret,
	}

	anyData := codectypes.UnsafePackAny(data)
	txData := &sdk.TxMsgData{
		MsgResponses: []*codectypes.Any{anyData},
	}

	txDataBz, err := proto.Marshal(txData)
	require.NoError(t, err)

	res, err := evmtypes.DecodeTxResponse(txDataBz)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, data.Logs, res.Logs)
	require.Equal(t, ret, res.Ret)
}

func TestUnwrapEthererumMsg(t *testing.T) {
	_, err := evmtypes.UnwrapEthereumMsg(nil, common.Hash{})
	require.NotNil(t, err)

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	clientCtx := client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	builder, _ := clientCtx.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)

	tx := builder.GetTx().(sdk.Tx)
	_, err = evmtypes.UnwrapEthereumMsg(&tx, common.Hash{})
	require.NotNil(t, err)

	evmTxParams := &evmtypes.EvmTxArgs{
		ChainID:  big.NewInt(1),
		Nonce:    0,
		To:       &common.Address{},
		Amount:   big.NewInt(0),
		GasLimit: 0,
		GasPrice: big.NewInt(0),
		Input:    []byte{},
	}

	msg := evmtypes.NewTx(evmTxParams)
	err = builder.SetMsgs(msg)
	require.Nil(t, err)

	tx = builder.GetTx().(sdk.Tx)
	unwrappedMsg, err := evmtypes.UnwrapEthereumMsg(&tx, msg.AsTransaction().Hash())
	require.Nil(t, err)
	require.Equal(t, unwrappedMsg, msg)
}

func TestBinSearch(t *testing.T) {
	successExecutable := func(gas uint64) (bool, *evmtypes.MsgEthereumTxResponse, error) {
		target := uint64(21000)
		return gas < target, nil, nil
	}
	failedExecutable := func(_ uint64) (bool, *evmtypes.MsgEthereumTxResponse, error) {
		return true, nil, errors.New("contract failed")
	}

	gas, err := evmtypes.BinSearch(20000, 21001, successExecutable)
	require.NoError(t, err)
	require.Equal(t, gas, uint64(21000))

	gas, err = evmtypes.BinSearch(20000, 21001, failedExecutable)
	require.Error(t, err)
	require.Equal(t, gas, uint64(0))
}

func TestTransactionLogsEncodeDecode(t *testing.T) {
	addr := utiltx.GenerateAddress().String()

	txLogs := evmtypes.TransactionLogs{
		Hash: common.BytesToHash([]byte("tx_hash")).String(),
		Logs: []*evmtypes.Log{
			{
				Address:     addr,
				Topics:      []string{common.BytesToHash([]byte("topic")).String()},
				Data:        []byte("data"),
				BlockNumber: 1,
				TxHash:      common.BytesToHash([]byte("tx_hash")).String(),
				TxIndex:     1,
				BlockHash:   common.BytesToHash([]byte("block_hash")).String(),
				Index:       1,
				Removed:     false,
			},
		},
	}

	txLogsEncoded, encodeErr := evmtypes.EncodeTransactionLogs(&txLogs)
	require.Nil(t, encodeErr)

	txLogsEncodedDecoded, decodeErr := evmtypes.DecodeTransactionLogs(txLogsEncoded)
	require.Nil(t, decodeErr)
	require.Equal(t, txLogs, txLogsEncodedDecoded)
}

func TestConvert18To6DecimalsCoin(t *testing.T) {
	testCases := []struct {
		name string
		coin sdk.Coin
		exp  sdk.Coin
	}{
		{
			name: "decimal < 5: 1.4",
			coin: sdk.NewCoin(utils.BaseDenom, math.NewInt(14e11)),
			exp:  sdk.NewCoin(utils.BaseDenom, math.NewInt(1)),
		},
		{
			name: "decimal == 5: 1.5",
			coin: sdk.NewCoin(utils.BaseDenom, math.NewInt(15e11)),
			exp:  sdk.NewCoin(utils.BaseDenom, math.NewInt(1)),
		},
		{
			name: "decimal > 5: 1.9",
			coin: sdk.NewCoin(utils.BaseDenom, math.NewInt(19e11)),
			exp:  sdk.NewCoin(utils.BaseDenom, math.NewInt(1)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := types.Convert18To6DecimalsCoin(tc.coin)
			require.Equal(t, tc.exp, res)
		})
	}
}

func TestZeroExtraDecimalsBigInt(t *testing.T) {
	testCases := []struct {
		name string
		amt  *big.Int
		exp  *big.Int
	}{
		{
			name: "almost 1: 0.99999...",
			amt:  big.NewInt(999999999999),
			exp:  big.NewInt(0),
		},
		{
			name: "decimal < 5: 1.4",
			amt:  big.NewInt(14e11),
			exp:  big.NewInt(1e12),
		},
		{
			name: "decimal < 5: 1.499999999999",
			amt:  big.NewInt(1499999999999),
			exp:  big.NewInt(1e12),
		},
		{
			name: "decimal == 5: 1.5",
			amt:  big.NewInt(15e11),
			exp:  big.NewInt(1e12),
		},
		{
			name: "decimal > 5: 1.9",
			amt:  big.NewInt(19e11),
			exp:  big.NewInt(1e12),
		},
		{
			name: "1 wei",
			amt:  big.NewInt(1),
			exp:  big.NewInt(0),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := types.ZeroExtraDecimalsBigInt(tc.amt)
			require.Equal(t, tc.exp, res)
		})
	}
}
