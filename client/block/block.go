// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package block

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	var height string
	cmd := &cobra.Command{
		Use:   "block",
		Short: "Get a specific block persisted in the db. If height is not specified, defaults to the latest.",
		Long:  "Get a specific block persisted in the db. If height is not specified, defaults to the latest.\nThis command works only if no other process is using the db. Before using it, make sure to stop your node.\nIf you're using a custom home directory, specify it with the '--home' flag",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			// Bind flags to the Context's Viper so the app construction can set
			// options accordingly.
			serverCtx := server.GetServerContextFromCmd(cmd)
			return serverCtx.Viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			cfg := serverCtx.Config
			home := cfg.RootDir

			store, err := newStore(home, server.GetAppDBBackend(serverCtx.Viper))
			if err != nil {
				return fmt.Errorf("error while openning db: %w", err)
			}

			state, err := store.state()
			if err != nil {
				return fmt.Errorf("error while getting blockstore state: %w", err)
			}

			var reqHeight int64
			if height != "latest" {
				reqHeight, err = strconv.ParseInt(height, 10, 64)
				if err != nil {
					return errors.New("invalid height, please provide an integer")
				}
				if reqHeight > state.Height {
					return fmt.Errorf("invalid height, the latest height found in the db is %d, and you asked for %d", state.Height, reqHeight)
				}
			} else {
				reqHeight = state.Height
			}

			block, err := store.block(reqHeight)
			if err != nil {
				return fmt.Errorf("error while getting block with height %d: %w", reqHeight, err)
			}

			bz, err := json.Marshal(block)
			if err != nil {
				return fmt.Errorf("error while parsing block to JSON: %w", err)
			}

			cmd.Println(string(bz))
			return nil
		},
	}

	cmd.Flags().StringVar(&height, "height", "latest", "Block height to retrieve")
	return cmd
}
