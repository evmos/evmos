// Copyright Tharsis Labs Ltd.(Eidon-chain)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/Eidon-AI/eidon-chain/blob/main/LICENSE)

//go:build !rocksdb
// +build !rocksdb

package main

import (
	"github.com/spf13/cobra"
)

// ChangeSetCmd returns nil for builds without rocksdb
// When building with rocksdb, ChangeSetCmd returns a Cobra command
// for interacting with change sets (check the 'versiondb.go' file)
func ChangeSetCmd() *cobra.Command {
	return nil
}
