// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE
package backend

import (
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

// GetLogs returns all the logs from all the ethereum transactions in a block.
func (b *Backend) GetLogs(hash common.Hash) ([][]*ethtypes.Log, error) {
	resBlock, err := b.TendermintBlockByHash(hash)
	if err != nil {
		return nil, err
	}
	if resBlock == nil {
		return nil, errors.Errorf("block not found for hash %s", hash)
	}
	return b.GetLogsByHeight(&resBlock.Block.Header.Height)
}

// GetLogsByHeight returns all the logs from all the ethereum transactions in a block.
func (b *Backend) GetLogsByHeight(height *int64) ([][]*ethtypes.Log, error) {
	// NOTE: we query the state in case the tx result logs are not persisted after an upgrade.
	blockRes, err := b.TendermintBlockResultByNumber(height)
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
