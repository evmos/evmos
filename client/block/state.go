package block

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb/opt"
	tmstore "github.com/tendermint/tendermint/proto/tendermint/store"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

type stateStore struct {
	state dbm.DB
	block dbm.DB
}

func newStateStore(path string) (*stateStore, error) {
	state, err := dbm.NewGoLevelDBWithOpts("state", path, &opt.Options{ReadOnly: true})
	if err != nil {
		return nil, err
	}

	block, err := dbm.NewGoLevelDBWithOpts("blockstore", path, &opt.Options{ReadOnly: true})
	if err != nil {
		return nil, err
	}

	return &stateStore{
		state: state,
		block: block,
	}, nil
}

// LoadBlockMeta returns the BlockMeta for the given height.
// If no block is found for the given height, it returns nil.
func (bs *stateStore) loadBlockMeta(height int64) (*types.BlockMeta, error) {
	var pbbm = new(tmproto.BlockMeta)
	bz, err := bs.block.Get(calcBlockMetaKey(height))

	if err != nil {
		panic(err)
	}

	if len(bz) == 0 {
		return nil, nil
	}

	err = pbbm.Unmarshal(bz)
	if err != nil {
		return nil, fmt.Errorf("unmarshal to tmproto.BlockMeta: %w", err)
	}

	blockMeta, err := types.BlockMetaFromProto(pbbm)
	if err != nil {
		return nil, fmt.Errorf("error from proto blockMeta: %w", err)
	}

	return blockMeta, nil
}

func calcBlockMetaKey(height int64) []byte {
	return []byte(fmt.Sprintf("H:%v", height))
}

func (bs *stateStore) loadBlock(height int64) *types.Block {
	blockMeta, err := bs.loadBlockMeta(height)
	if err != nil {
		panic(err)
	}
	if blockMeta == nil {
		return nil
	}

	pbb := new(tmproto.Block)
	buf := []byte{}
	for i := 0; i < int(blockMeta.BlockID.PartSetHeader.Total); i++ {
		part := bs.loadBlockPart(height, i)
		// If the part is missing (e.g. since it has been deleted after we
		// loaded the block meta) we consider the whole block to be missing.
		if part == nil {
			return nil
		}
		buf = append(buf, part.Bytes...)
	}
	err = pbb.Unmarshal(buf)
	if err != nil {
		// NOTE: The existence of meta should imply the existence of the
		// block. So, make sure meta is only saved after blocks are saved.
		panic(fmt.Sprintf("Error reading block: %v", err))
	}

	block, err := types.BlockFromProto(pbb)
	if err != nil {
		panic(fmt.Errorf("error from proto block: %w", err))
	}

	return block
}

func (bs *stateStore) loadBlockPart(height int64, index int) *types.Part {
	var pbpart = new(tmproto.Part)

	bz, err := bs.block.Get(calcBlockPartKey(height, index))
	if err != nil {
		panic(err)
	}
	if len(bz) == 0 {
		return nil
	}

	err = pbpart.Unmarshal(bz)
	if err != nil {
		panic(fmt.Errorf("unmarshal to tmproto.Part failed: %w", err))
	}
	part, err := types.PartFromProto(pbpart)
	if err != nil {
		panic(fmt.Sprintf("Error reading block part: %v", err))
	}

	return part
}

func calcBlockPartKey(height int64, partIndex int) []byte {
	return []byte(fmt.Sprintf("P:%v:%v", height, partIndex))
}

// LoadBlockStoreState returns the BlockStoreState as loaded from disk.
// If no BlockStoreState was previously persisted, it returns the zero value.
var blockStoreKey = []byte("blockStore")

func (bs *stateStore) loadBlockStoreState() (base int64, height int64) {
	bytes, err := bs.block.Get(blockStoreKey)
	if err != nil {
		panic(err)
	}

	if len(bytes) == 0 {
		return 0, 0
	}

	var bsj tmstore.BlockStoreState
	if err := bsj.Unmarshal(bytes); err != nil {
		panic(fmt.Sprintf("Could not unmarshal bytes: %X", bytes))
	}

	// Backwards compatibility with persisted data from before Base existed.
	if bsj.Height > 0 && bsj.Base == 0 {
		bsj.Base = 1
	}

	return bsj.Base, bsj.Height
}
