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

package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/evmos/evmos/v12/x/vesting/types"
)

// GetQueryCmd returns the parent command for all vesting CLI query commands.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the vesting module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetBalancesCmd(),
	)
	return cmd
}

// GetBalancesCmd queries the unvested tokens for a given vesting account
func GetBalancesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balances ADDRESS",
		Short: "Gets locked, unvested and vested tokens for a vesting account",
		Long:  "Gets locked, unvested and vested tokens for a vesting account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryBalancesRequest{
				Address: args[0],
			}

			res, err := queryClient.Balances(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(
				fmt.Sprintf("Locked: %s\nUnvested: %s\nVested: %s\n", res.Locked, res.Unvested, res.Vested))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
