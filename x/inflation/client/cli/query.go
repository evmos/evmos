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

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/evmos/evmos/v10/x/inflation/types"
)

// GetQueryCmd returns the cli query commands for the inflation module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the inflation module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetPeriod(),
		GetEpochMintProvision(),
		GetSkippedEpochs(),
		GetCirculatingSupply(),
		GetInflationRate(),
		GetParams(),
	)

	return cmd
}

// GetPeriod implements a command to return the current inflation
// period
func GetPeriod() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "period",
		Short: "Query the current inflation period",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryPeriodRequest{}
			res, err := queryClient.Period(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%v\n", res.Period))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetEpochMintProvision implements a command to return the current inflation
// epoch provisions value.
func GetEpochMintProvision() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "epoch-mint-provision",
		Short: "Query the current inflation epoch provisions value",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryEpochMintProvisionRequest{}
			res, err := queryClient.EpochMintProvision(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%s\n", res.EpochMintProvision))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetSkippedEpochs implements a command to return the current inflation
// period
func GetSkippedEpochs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skipped-epochs",
		Short: "Query the current number of skipped epochs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QuerySkippedEpochsRequest{}
			res, err := queryClient.SkippedEpochs(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%v\n", res.SkippedEpochs))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCirculatingSupply implements a command to return the current circulating supply
func GetCirculatingSupply() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "circulating-supply",
		Short: "Query the current supply of tokens in circulation",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryCirculatingSupplyRequest{}
			res, err := queryClient.CirculatingSupply(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%s\n", res.CirculatingSupply))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetInflationRate implements a command to return the inflation rate in %
func GetInflationRate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inflation-rate",
		Short: "Query the inflation rate of the current period",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryInflationRateRequest{}
			res, err := queryClient.InflationRate(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%s%%\n", res.InflationRate))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetParams implements a command to return the current inflation
// parameters.
func GetParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current inflation parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryParamsRequest{}
			res, err := queryClient.Params(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
