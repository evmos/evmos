//go:build !rocksdb
// +build !rocksdb

package opendb

import (
	"path/filepath"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cosmos/cosmos-sdk/server/types"
)

func OpenDB(_ types.AppOptions, home string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(home, "data")
	return dbm.NewDB("application", backendType, dataDir)
}

// OpenReadOnlyDB opens rocksdb backend in read-only mode.
func OpenReadOnlyDB(home string, backendType dbm.BackendType) (dbm.DB, error) {
	return OpenDB(nil, home, backendType)
}
