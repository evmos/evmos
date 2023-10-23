// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package types

import (
	abci "github.com/cometbft/cometbft/abci/types"
	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/ethereum/go-ethereum/common"
)

// EVMTxIndexer defines the interface of custom eth tx indexer.
type EVMTxIndexer interface {
	// LastIndexedBlock returns -1 if indexer db is empty
	LastIndexedBlock() (int64, error)
	IndexBlock(*tmtypes.Block, []*abci.ResponseDeliverTx) error

	// GetByTxHash returns nil if tx not found.
	GetByTxHash(common.Hash) (*TxResult, error)
	// GetByBlockAndIndex returns nil if tx not found.
	GetByBlockAndIndex(int64, int32) (*TxResult, error)
}

// IsEthTx check if the tx is an eth tx
func IsEthTx(tx sdk.Tx) bool {
	extTx, ok := tx.(authante.HasExtensionOptionsTx)
	if !ok {
		return false
	}
	opts := extTx.GetExtensionOptions()
	if len(opts) != 1 || opts[0].GetTypeUrl() != "/ethermint.evm.v1.ExtensionOptionsEthereumTx" {
		return false
	}
	return true
}
