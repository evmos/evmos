// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package block

import (
	"fmt"
	"path/filepath"

	dbm "github.com/cometbft/cometbft-db"
	tmstore "github.com/cometbft/cometbft/proto/tendermint/store"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/types"
	"github.com/cosmos/gogoproto/proto"
)

var blockStoreKey = []byte("blockStore")

type blockStore struct {
	dbm.DB
}

// newBlockStore opens the 'blockstore' db
// and returns it.
func newBlockStore(rootDir string, backendType dbm.BackendType) (*blockStore, error) {
	dataDir := filepath.Join(rootDir, "data")
	store, err := dbm.NewDB("blockstore", backendType, dataDir)
	if err != nil {
		return nil, err
	}

	return &blockStore{store}, nil
}

// state returns the BlockStoreState as loaded from disk.
// If no BlockStoreState was previously persisted, it returns nil.
func (bs *blockStore) state() *tmstore.BlockStoreState {
	bytes, err := bs.Get(blockStoreKey)
	if err != nil {
		panic(err)
	}

	if len(bytes) == 0 {
		return nil
	}

	var bsj tmstore.BlockStoreState
	if err := proto.Unmarshal(bytes, &bsj); err != nil {
		panic(fmt.Sprintf("could not unmarshal bytes: %X", bytes))
	}

	// Backwards compatibility with persisted data from before Base existed.
	if bsj.Height > 0 && bsj.Base == 0 {
		bsj.Base = 1
	}

	return &bsj
}

// block returns the Block for the given height.
// If no block is found for the given height, it returns nil.
func (bs *blockStore) block(height int64) *types.Block {
	blockMeta, err := bs.blockMeta(height)
	if err != nil {
		panic(err)
	}
	if blockMeta == nil {
		return nil
	}

	pbb := new(tmproto.Block)
	buf := []byte{}
	for i := 0; i < int(blockMeta.BlockID.PartSetHeader.Total); i++ {
		part := bs.blockPart(height, i)
		// If the part is missing (e.g. since it has been deleted after we
		// loaded the block meta) we consider the whole block to be missing.
		if part == nil {
			return nil
		}
		buf = append(buf, part.Bytes...)
	}
	if err := proto.Unmarshal(buf, pbb); err != nil {
		// NOTE: The existence of meta should imply the existence of the
		// block. So, make sure meta is only saved after blocks are saved.
		panic(fmt.Sprintf("error reading block: %v", err))
	}

	block, err := types.BlockFromProto(pbb)
	if err != nil {
		panic(fmt.Errorf("error from proto block: %w", err))
	}

	return block
}

// blockMeta returns the BlockMeta for the given height.
// If no block is found for the given height, it returns nil.
func (bs *blockStore) blockMeta(height int64) (*types.BlockMeta, error) {
	bz, err := bs.Get(blockMetaKey(height))
	if err != nil {
		panic(err)
	}

	if len(bz) == 0 {
		return nil, nil
	}

	pbbm := new(tmproto.BlockMeta)
	err = proto.Unmarshal(bz, pbbm)
	if err != nil {
		return nil, fmt.Errorf("unmarshal to tmproto.BlockMeta: %w", err)
	}

	blockMeta, err := types.BlockMetaFromProto(pbbm)
	if err != nil {
		return nil, fmt.Errorf("error from proto blockMeta: %w", err)
	}

	return blockMeta, nil
}

// blockPart returns the part of the block for the given height and part index.
// If no block part is found for the given height and index, it returns nil.
func (bs *blockStore) blockPart(height int64, index int) *types.Part {
	bz, err := bs.Get(blockPartKey(height, index))
	if err != nil {
		panic(err)
	}
	if len(bz) == 0 {
		return nil
	}

	pbpart := new(tmproto.Part)
	if err := proto.Unmarshal(bz, pbpart); err != nil {
		panic(fmt.Errorf("unmarshal to tmproto.Part failed: %w", err))
	}
	part, err := types.PartFromProto(pbpart)
	if err != nil {
		panic(fmt.Sprintf("error reading block part: %v", err))
	}

	return part
}

// blockMetaKey is a helper function that takes the block height
// as input parameter and returns the corresponding block metadata store key
func blockMetaKey(height int64) []byte {
	return []byte(fmt.Sprintf("H:%v", height))
}

// blockPartKey is a helper function that takes the block height
// and the part index as input parameters and returns the corresponding block part store key
func blockPartKey(height int64, partIndex int) []byte {
	return []byte(fmt.Sprintf("P:%v:%v", height, partIndex))
}
