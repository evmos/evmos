package block

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func LastBlockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "last-block [path_to_db]",
		Short: "Get the last block of the db",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO add flag to specify different dbs
			statedb, err := newStateStore(args[0])
			if err != nil {
				return fmt.Errorf("error while openning db: %w", err)
			}

			blockStore := statedb.loadBlockStoreState()
			if blockStore == nil {
				return errors.New("couldn't find a BlockStoreState persisted in db")
			}
			block := statedb.loadBlock(blockStore.Height)

			bz, err := json.Marshal(block)
			if err != nil {
				return err
			}

			fmt.Println(string(bz))

			return nil
		},
	}
	return cmd
}
