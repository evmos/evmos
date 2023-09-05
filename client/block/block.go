package block

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/types"
)

func LastBlockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "last-block [path_to_db]",
		Short: "Get the base and highest height of the db",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			block, err := getLatestBlock(args[0])
			if err != nil {
				return err
			}

			blockStr, err := json.Marshal(block)
			if err != nil {
				return err
			}

			fmt.Println(string(blockStr))

			return nil
		},
	}
	return cmd
}

func getLatestBlock(path string) (*types.Block, error) {
	statedb, err := newStateStore(path)
	if err != nil {
		return nil,fmt.Errorf("new stateStore: %w", err)
	}
	_, height := statedb.loadBlockStoreState()
	block := statedb.loadBlock(height)

	return block, nil
}