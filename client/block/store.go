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

var storeKey = []byte("blockStore")

// store is the block store struct
type store struct {
	dbm.DB
}

// newStore opens the 'blockstore' db
// and returns it.
func newStore(rootDir string, backendType dbm.BackendType) (*store, error) {
	dataDir := filepath.Join(rootDir, "data")
	db, err := dbm.NewDB("blockstore", backendType, dataDir)
	if err != nil {
		return nil, err
	}

	return &store{db}, nil
}

// state returns the BlockStoreState as loaded from disk.
// If no BlockStoreState was previously persisted, it returns nil.
func (s *store) state() *tmstore.BlockStoreState {
	bytes, err := s.Get(storeKey)
	if err != nil {
		panic(err)
	}

	if len(bytes) == 0 {
		return nil
	}

	var bss tmstore.BlockStoreState
	if err := proto.Unmarshal(bytes, &bss); err != nil {
		panic(fmt.Sprintf("could not unmarshal bytes: %X", bytes))
	}

	// Backwards compatibility with persisted data from before Base existed.
	if bss.Height > 0 && bss.Base == 0 {
		bss.Base = 1
	}

	return &bss
}

// block returns the Block for the given height.
// If no block is found for the given height, it returns nil.
func (s *store) block(height int64) *types.Block {
	bm, err := s.meta(height)
	if err != nil {
		panic(err)
	}
	if bm == nil {
		return nil
	}

	pbb := new(tmproto.Block)
	buf := []byte{}
	for i := 0; i < int(bm.BlockID.PartSetHeader.Total); i++ {
		part := s.part(height, i)
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

	b, err := types.BlockFromProto(pbb)
	if err != nil {
		panic(fmt.Errorf("error from proto block: %w", err))
	}

	return b
}

// meta returns the BlockMeta for the given height.
// If no block is found for the given height, it returns nil.
func (s *store) meta(height int64) (*types.BlockMeta, error) {
	bz, err := s.Get(metaKey(height))
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

	bm, err := types.BlockMetaFromProto(pbbm)
	if err != nil {
		return nil, fmt.Errorf("error from proto blockMeta: %w", err)
	}

	return bm, nil
}

// part returns the part of the block for the given height and part index.
// If no block part is found for the given height and index, it returns nil.
func (s *store) part(height int64, index int) *types.Part {
	bz, err := s.Get(partKey(height, index))
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

// metaKey is a helper function that takes the block height
// as input parameter and returns the corresponding block metadata store key
func metaKey(height int64) []byte {
	return []byte(fmt.Sprintf("H:%v", height))
}

// partKey is a helper function that takes the block height
// and the part index as input parameters and returns the corresponding block part store key
func partKey(height int64, partIndex int) []byte {
	return []byte(fmt.Sprintf("P:%v:%v", height, partIndex))
}
