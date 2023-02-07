// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

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

	"github.com/evmos/evmos/v11/utils"
)

// FlagGenesisTime defines the genesis time in string format
const FlagGenesisTime = "genesis-time"

var migrationMap = genutiltypes.MigrationMap{}

// GetMigrationCallback returns a MigrationCallback for a given version.
func GetMigrationCallback(version, chainID string) genutiltypes.MigrationCallback {
	if !utils.IsMainnet(chainID) {
		version = fmt.Sprintf("%s%s", "t", version)
	}

	return migrationMap[version]
}

// MigrateGenesisCmd returns a command to execute genesis state migration.
func MigrateGenesisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate TARGET_VERSION GENESIS_FILE",
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
