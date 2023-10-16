// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

//go:build !rocksdb
// +build !rocksdb

package app

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// versionDB constant for 'versiondb'
// is same constant as in 'app/db.go' but need to include it here too
// cause only one of these files (db.go or db_placeholder.go) will be
// included in the compiled binary depending on the build type (with or without rocksdb)
const versionDB = "versiondb"

// setupVersionDB returns error on non-rocksdb build
// because it is not supported in other builds
// If you're building the binary with rocksdb,
// the setupVersionDB function from the 'app/db.go' file
// will be called
func setupVersionDB(
	_ string,
	_ *baseapp.BaseApp,
	_ map[string]*storetypes.KVStoreKey,
	_ map[string]*storetypes.TransientStoreKey,
	_ map[string]*storetypes.MemoryStoreKey,
) (sdk.MultiStore, error) {
	return nil, errors.New("versiondb is not supported in this binary")
}
