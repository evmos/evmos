package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	utiltx "github.com/evmos/evmos/v14/testutil/tx"
	"github.com/evmos/evmos/v14/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
)

func TestTransactionLogsValidate(t *testing.T) {
	addr := utiltx.GenerateAddress().String()

	testCases := []struct {
		name    string
		txLogs  types.TransactionLogs
		expPass bool
	}{
		{
			"valid log",
			types.TransactionLogs{
				Hash: common.BytesToHash([]byte("tx_hash")).String(),
				Logs: []*types.Log{
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
			},
			true,
		},
		{
			"empty hash",
			types.TransactionLogs{
				Hash: common.Hash{}.String(),
			},
			false,
		},
		{
			"nil log",
			types.TransactionLogs{
				Hash: common.BytesToHash([]byte("tx_hash")).String(),
				Logs: []*types.Log{nil},
			},
			false,
		},
		{
			"invalid log",
			types.TransactionLogs{
				Hash: common.BytesToHash([]byte("tx_hash")).String(),
				Logs: []*types.Log{{}},
			},
			false,
		},
		{
			"hash mismatch log",
			types.TransactionLogs{
				Hash: common.BytesToHash([]byte("tx_hash")).String(),
				Logs: []*types.Log{
					{
						Address:     addr,
						Topics:      []string{common.BytesToHash([]byte("topic")).String()},
						Data:        []byte("data"),
						BlockNumber: 1,
						TxHash:      common.BytesToHash([]byte("other_hash")).String(),
						TxIndex:     1,
						BlockHash:   common.BytesToHash([]byte("block_hash")).String(),
						Index:       1,
						Removed:     false,
					},
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		err := tc.txLogs.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}

func TestValidateLog(t *testing.T) {
	addr := utiltx.GenerateAddress().String()

	testCases := []struct {
		name    string
		log     *types.Log
		expPass bool
	}{
		{
			"valid log",
			&types.Log{
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
			true,
		},
		{
			"empty log", &types.Log{}, false,
		},
		{
			"zero address",
			&types.Log{
				Address: common.Address{}.String(),
			},
			false,
		},
		{
			"empty block hash",
			&types.Log{
				Address:   addr,
				BlockHash: common.Hash{}.String(),
			},
			false,
		},
		{
			"zero block number",
			&types.Log{
				Address:     addr,
				BlockHash:   common.BytesToHash([]byte("block_hash")).String(),
				BlockNumber: 0,
			},
			false,
		},
		{
			"empty tx hash",
			&types.Log{
				Address:     addr,
				BlockHash:   common.BytesToHash([]byte("block_hash")).String(),
				BlockNumber: 1,
				TxHash:      common.Hash{}.String(),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		err := tc.log.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}

func TestConversionFunctions(t *testing.T) {
	addr := utiltx.GenerateAddress().String()

	txLogs := types.TransactionLogs{
		Hash: common.BytesToHash([]byte("tx_hash")).String(),
		Logs: []*types.Log{
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

	// convert valid log to eth logs and back (and validate)
	conversionLogs := types.NewTransactionLogsFromEth(common.BytesToHash([]byte("tx_hash")), txLogs.EthLogs())
	conversionErr := conversionLogs.Validate()

	// create new transaction logs as copy of old valid one (and validate)
	copyLogs := types.NewTransactionLogs(common.BytesToHash([]byte("tx_hash")), txLogs.Logs)
	copyErr := copyLogs.Validate()

	require.Nil(t, conversionErr)
	require.Nil(t, copyErr)
}
