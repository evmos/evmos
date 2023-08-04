// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package statedb

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/evmos/evmos/v14/x/evm/types"
)

// TxConfig encapulates the readonly information of current tx for `StateDB`.
type TxConfig struct {
	BlockHash common.Hash // hash of current block
	TxHash    common.Hash // hash of current tx
	TxIndex   uint        // the index of current transaction
	LogIndex  uint        // the index of next log within current block
}

// NewTxConfig returns a TxConfig
func NewTxConfig(bhash, thash common.Hash, txIndex, logIndex uint) TxConfig {
	return TxConfig{
		BlockHash: bhash,
		TxHash:    thash,
		TxIndex:   txIndex,
		LogIndex:  logIndex,
	}
}

// NewEmptyTxConfig construct an empty TxConfig,
// used in context where there's no transaction, e.g. `eth_call`/`eth_estimateGas`.
func NewEmptyTxConfig(bhash common.Hash) TxConfig {
	return TxConfig{
		BlockHash: bhash,
		TxHash:    common.Hash{},
		TxIndex:   0,
		LogIndex:  0,
	}
}

// EVMConfig encapsulates common parameters needed to create an EVM to execute a message
// It's mainly to reduce the number of method parameters
type EVMConfig struct {
	Params      types.Params
	ChainConfig *params.ChainConfig
	CoinBase    common.Address
	BaseFee     *big.Int
}
