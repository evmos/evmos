// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

//go:build !rocksdb
// +build !rocksdb

package main

import (
	"github.com/spf13/cobra"
)

func ChangeSetCmd() *cobra.Command {
	return nil
}
