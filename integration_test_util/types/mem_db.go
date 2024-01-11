package types

import (
	cdb "github.com/cometbft/cometbft-db"
	// tdb "github.com/tendermint/tm-db"
)

var _ cdb.DB = (*MemDB)(nil)

// MemDB is a wrapper of Tendermint/CometBFT DB that is backward-compatible with CometBFT chains pre-rename package.
//
// (eg: replace github.com/tendermint/tendermint => github.com/cometbft/cometbft v0.34.29)
type MemDB struct {
	cDb cdb.DB
	// tmDb tdb.DB
}

//func WrapTendermintDB(tmDb tdb.DB) *MemDB {
//	return &MemDB{tmDb: tmDb}
//}

func WrapCometBftDB(cDb cdb.DB) *MemDB {
	return &MemDB{cDb: cDb}
}

func (w *MemDB) AsCometBFT() cdb.DB {
	return w
}

//func (w *MemDB) AsTendermint() tdb.DB {
//	return w.tmDb
//}

func (w *MemDB) Get(bytes []byte) ([]byte, error) {
	return w.cDb.Get(bytes)
}

func (w *MemDB) Has(key []byte) (bool, error) {
	return w.cDb.Has(key)
}

func (w *MemDB) Set(bytes []byte, bytes2 []byte) error {
	return w.cDb.Set(bytes, bytes2)
}

func (w *MemDB) SetSync(bytes []byte, bytes2 []byte) error {
	return w.cDb.SetSync(bytes, bytes2)
}

func (w *MemDB) Delete(bytes []byte) error {
	return w.cDb.Delete(bytes)
}

func (w *MemDB) DeleteSync(bytes []byte) error {
	return w.cDb.DeleteSync(bytes)
}

func (w *MemDB) Iterator(start, end []byte) (cdb.Iterator, error) {
	return w.cDb.Iterator(start, end)
}

func (w *MemDB) ReverseIterator(start, end []byte) (cdb.Iterator, error) {
	return w.cDb.ReverseIterator(start, end)
}

func (w *MemDB) Close() error {
	return w.cDb.Close()
}

func (w *MemDB) NewBatch() cdb.Batch {
	return w.cDb.NewBatch()
}

func (w *MemDB) Print() error {
	return w.cDb.Print()
}

func (w *MemDB) Stats() map[string]string {
	return w.cDb.Stats()
}
