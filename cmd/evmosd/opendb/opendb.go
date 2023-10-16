// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

//go:build !rocksdb
// +build !rocksdb

package opendb

import (
	"path/filepath"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cosmos/cosmos-sdk/server/types"
)

// OpenDB opens a database based on the specified backend type.
// It takes the home directory where the database data will be stored, along with the backend type.
// It opens a database named "application" using the specified backend type and the data directory.
// It returns the opened database and an error (if any). If the database opens successfully, the error will be nil.
// NOTE: this is included in builds without rocksdb.
// When building the binary with rocksdb, the code in 'rocksdb.go' will be included
// instead of this
func OpenDB(_ types.AppOptions, home string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(home, "data")
	return dbm.NewDB("application", backendType, dataDir)
}

// OpenReadOnlyDB opens rocksdb backend in read-only mode.
func OpenReadOnlyDB(home string, backendType dbm.BackendType) (dbm.DB, error) {
	return OpenDB(nil, home, backendType)
}
