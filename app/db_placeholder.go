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

const versionDB = "versiondb"

func setupVersionDB(
	_ string,
	_ *baseapp.BaseApp,
	_ map[string]*storetypes.KVStoreKey,
	_ map[string]*storetypes.TransientStoreKey,
	_ map[string]*storetypes.MemoryStoreKey,
) (sdk.MultiStore, error) {
	return nil, errors.New("versiondb is not supported in this binary")
}
