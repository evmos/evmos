// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package v2

import (
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

// GetLogs returns all the logs from all the ethereum transactions in a block.
func (b *Backend) GetLogs(hash common.Hash) ([][]*ethtypes.Log, error) {
	resHeader, err := b.rpcClient.HeaderByHash(b.ctx, hash.Bytes())
	if err != nil {
		return nil, err
	}

	if resHeader == nil {
		return nil, errors.Errorf("header not found for hash %s", hash)
	}

	return b.GetLogsByHeight(&resHeader.Header.Height)
}

// GetLogsByHeight returns all the logs from all the ethereum transactions in a block.
func (b *Backend) GetLogsByHeight(height *int64) ([][]*ethtypes.Log, error) {
	// NOTE: we query the state in case the tx result logs are not persisted after an upgrade.
	blockRes, err := b.rpcClient.BlockResults(b.ctx, height)
	if err != nil {
		return nil, err
	}

	return GetLogsFromBlockResults(blockRes)
}

// BloomStatus returns the BloomBitsBlocks and the number of processed sections maintained
// by the chain indexer.
func (b *Backend) BloomStatus() (uint64, uint64) {
	return 4096, 0
}
