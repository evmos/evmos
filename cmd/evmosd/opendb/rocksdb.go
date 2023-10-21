// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

//go:build rocksdb
// +build rocksdb

package opendb

import (
	"path/filepath"
	"runtime"
	"strings"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/linxGnu/grocksdb"

	"github.com/cosmos/cosmos-sdk/server/types"
)

// BlockCacheSize defines the size of the cache of a block.
const BlockCacheSize = 1 << 30

// OpenDB opens a database based on the specified backend type.
// It takes the home directory where the database data will be stored, along with the backend type.
// If the backend type is RocksDB, it opens a RocksDB database located in a subdirectory called "application.db" under the data directory.
// For other backend types, it opens a database named "application" using the specified backend type and the data directory.
// It returns the opened database and an error (if any). If the database opens successfully, the error will be nil.
// NOTE: This function is only included in builds with rocksdb.
// Otherwise, the 'OpenDB' function from 'opendb.go' will be included
func OpenDB(_ types.AppOptions, home string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(home, "data")
	if backendType == dbm.RocksDBBackend {
		return openRocksdb(filepath.Join(dataDir, "application.db"), false)
	}

	return dbm.NewDB("application", backendType, dataDir)
}

// OpenReadOnlyDB opens rocksdb backend in read-only mode.
func OpenReadOnlyDB(home string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(home, "data")
	if backendType == dbm.RocksDBBackend {
		return openRocksdb(filepath.Join(dataDir, "application.db"), true)
	}

	return dbm.NewDB("application", backendType, dataDir)
}

func openRocksdb(dir string, readonly bool) (dbm.DB, error) {
	opts, err := loadLatestOptions(dir)
	if err != nil {
		return nil, err
	}
	// customize rocksdb options
	opts = NewRocksdbOptions(opts, false)

	var db *grocksdb.DB
	if readonly {
		db, err = grocksdb.OpenDbForReadOnly(opts, dir, false)
	} else {
		db, err = grocksdb.OpenDb(opts, dir)
	}
	if err != nil {
		return nil, err
	}

	ro := grocksdb.NewDefaultReadOptions()
	wo := grocksdb.NewDefaultWriteOptions()
	woSync := grocksdb.NewDefaultWriteOptions()
	woSync.SetSync(true)
	return dbm.NewRocksDBWithRawDB(db, ro, wo, woSync), nil
}

// loadLatestOptions try to load options from existing db, returns nil if not exists.
func loadLatestOptions(dir string) (*grocksdb.Options, error) {
	opts, err := grocksdb.LoadLatestOptions(dir, grocksdb.NewDefaultEnv(), true, grocksdb.NewLRUCache(BlockCacheSize))
	if err != nil {
		// not found is not an error
		if strings.HasPrefix(err.Error(), "NotFound: ") {
			return nil, nil
		}
		return nil, err
	}

	cfNames := opts.ColumnFamilyNames()
	cfOpts := opts.ColumnFamilyOpts()

	for i := 0; i < len(cfNames); i++ {
		if cfNames[i] == "default" {
			return &cfOpts[i], nil
		}
	}

	return opts.Options(), nil
}

// NewRocksdbOptions build options for `application.db`,
// it overrides existing options if provided, otherwise create new one assuming it's a new database.
func NewRocksdbOptions(opts *grocksdb.Options, sstFileWriter bool) *grocksdb.Options {
	if opts == nil {
		opts = grocksdb.NewDefaultOptions()
		// only enable dynamic-level-bytes on new db, don't override for existing db
		opts.SetLevelCompactionDynamicLevelBytes(true)
	}
	opts.SetCreateIfMissing(true)
	opts.IncreaseParallelism(runtime.NumCPU())
	opts.OptimizeLevelStyleCompaction(512 * 1024 * 1024)
	opts.SetTargetFileSizeMultiplier(2)

	// block based table options
	bbto := grocksdb.NewDefaultBlockBasedTableOptions()

	// 1G block cache
	bbto.SetBlockCache(grocksdb.NewLRUCache(BlockCacheSize))

	// http://rocksdb.org/blog/2021/12/29/ribbon-filter.html
	bbto.SetFilterPolicy(grocksdb.NewRibbonHybridFilterPolicy(9.9, 1))

	// partition index
	// http://rocksdb.org/blog/2017/05/12/partitioned-index-filter.html
	bbto.SetIndexType(grocksdb.KTwoLevelIndexSearchIndexType)
	bbto.SetPartitionFilters(true)

	// hash index is better for iavl tree which mostly do point lookup.
	bbto.SetDataBlockIndexType(grocksdb.KDataBlockIndexTypeBinarySearchAndHash)

	opts.SetBlockBasedTableFactory(bbto)

	// in iavl tree, we almost always query existing keys
	opts.SetOptimizeFiltersForHits(true)

	// heavier compression option at bottommost level,
	// 110k dict bytes is default in zstd library,
	// train bytes is recommended to be set at 100x dict bytes.
	opts.SetBottommostCompression(grocksdb.ZSTDCompression)
	compressOpts := grocksdb.NewDefaultCompressionOptions()
	compressOpts.Level = 12
	if !sstFileWriter {
		compressOpts.MaxDictBytes = 110 * 1024
		opts.SetBottommostCompressionOptionsZstdMaxTrainBytes(compressOpts.MaxDictBytes*100, true)
	}
	opts.SetBottommostCompressionOptions(compressOpts, true)
	return opts
}
