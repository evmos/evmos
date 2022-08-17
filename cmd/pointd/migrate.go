package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	v3 "github.com/evmos/evmos/v8/app/upgrades/v3"
	v5 "github.com/evmos/evmos/v8/app/upgrades/v5"
	"github.com/evmos/evmos/v8/types"
)

// FlagGenesisTime defines the genesis time in string format
const FlagGenesisTime = "genesis-time"

var migrationMap = genutiltypes.MigrationMap{
	"v3": v3.MigrateGenesis, // migration to v3
	"v5": v5.MigrateGenesis, // migration to v5
}

// GetMigrationCallback returns a MigrationCallback for a given version.
func GetMigrationCallback(version, chainID string) genutiltypes.MigrationCallback {
	if !types.IsMainnet(chainID) {
		version = fmt.Sprintf("%s%s", "t", version)
	}

	return migrationMap[version]
}

// MigrateGenesisCmd returns a command to execute genesis state migration.
func MigrateGenesisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate [target-version] [genesis-file]",
		Short: "Migrate genesis to a specified target version",
		Long:  "Migrate the source genesis into the target version and print to STDOUT.",
		Example: fmt.Sprintf(
			"%s migrate v3 /path/to/genesis.json --chain-id=evmos_9001-2 --genesis-time=2022-04-01T17:00:00Z",
			version.AppName,
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			target := args[0]
			importGenesis := args[1]

			genDoc, err := tmtypes.GenesisDocFromFile(importGenesis)
			if err != nil {
				return fmt.Errorf("failed to retrieve genesis.json: %w", err)
			}

			var initialState genutiltypes.AppMap
			if err := json.Unmarshal(genDoc.AppState, &initialState); err != nil {
				return fmt.Errorf("failed to JSON unmarshal initial genesis state: %w", err)
			}

			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			if chainID != "" {
				genDoc.ChainID = chainID
			}

			migrationFn := GetMigrationCallback(target, chainID)
			if migrationFn == nil {
				return fmt.Errorf("unknown migration function for version: %s", target)
			}

			newGenState := migrationFn(initialState, clientCtx)

			appState, err := json.Marshal(newGenState)
			if err != nil {
				return fmt.Errorf("failed to JSON marshal migrated genesis state: %w", err)
			}

			genDoc.AppState = appState

			genesisTime, _ := cmd.Flags().GetString(FlagGenesisTime)
			if genesisTime != "" {
				var t time.Time

				if err := t.UnmarshalText([]byte(genesisTime)); err != nil {
					return fmt.Errorf("failed to unmarshal genesis time: %w", err)
				}

				genDoc.GenesisTime = t
			}

			bz, err := tmjson.Marshal(genDoc)
			if err != nil {
				return fmt.Errorf("failed to marshal genesis doc: %w", err)
			}

			sortedBz, err := sdk.SortJSON(bz)
			if err != nil {
				return fmt.Errorf("failed to sort JSON genesis doc: %w", err)
			}

			cmd.Println(string(sortedBz))
			return nil
		},
	}

	cmd.Flags().String(FlagGenesisTime, "", "override genesis time")
	cmd.Flags().String(flags.FlagChainID, "", "override genesis chain-id")

	return cmd
}
