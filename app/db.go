// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

//go:build rocksdb
// +build rocksdb

package app

import (
	"os"
	"path/filepath"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/crypto-org-chain/cronos/versiondb"
	"github.com/crypto-org-chain/cronos/versiondb/tsrocksdb"
)

// versionDB constant for 'versiondb'
// is same constant as in 'app/db_placeholder.go' but need to include it here too
// cause only one of these files ('db.go' or 'db_placeholder.go') will be
// included in the compiled binary depending on the build type (with or without rocksdb)
const versionDB = "versiondb"

// setupVersionDB sets up versionDB and
// returns the corresponding QueryMultiStore
// NOTE: this code is only included in a build with rocksdb.
// Otherwise, the setupVersionDB code on 'app/db_placeholder.go' will be included
// in the compiled binary
func (app *Evmos) setupVersionDB(
	homePath string,
	keys map[string]*storetypes.KVStoreKey,
	tkeys map[string]*storetypes.TransientStoreKey,
	memKeys map[string]*storetypes.MemoryStoreKey,
) (sdk.MultiStore, error) {
	dataDir := filepath.Join(homePath, "data", versionDB)
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		return nil, err
	}
	store, err := tsrocksdb.NewStore(dataDir)
	if err != nil {
		return nil, err
	}

	// default to exposing all
	exposeStoreKeys := make([]storetypes.StoreKey, 0, len(keys))
	for _, storeKey := range keys {
		exposeStoreKeys = append(exposeStoreKeys, storeKey)
	}

	service := versiondb.NewStreamingService(store, exposeStoreKeys)
	app.SetStreamingService(service)

	verDB := versiondb.NewMultiStore(app.CommitMultiStore(), store, exposeStoreKeys)
	verDB.MountTransientStores(tkeys)
	verDB.MountMemoryStores(memKeys)

	app.SetQueryMultiStore(verDB)
	return verDB, nil
}
