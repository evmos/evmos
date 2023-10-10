//go:build !rocksdb
// +build !rocksdb

package main

import (
	"github.com/spf13/cobra"
)

func ChangeSetCmd() *cobra.Command {
	return nil
}
