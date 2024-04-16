// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/evmos/evmos/v18/x/inflation/v1/types"
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
		RunE: func(cmd *cobra.Command, _ []string) error {
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryEpochMintProvisionRequest{}
			res, err := queryClient.EpochMintProvision(context.Background(), req)
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QuerySkippedEpochsRequest{}
			res, err := queryClient.SkippedEpochs(context.Background(), req)
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryCirculatingSupplyRequest{}
			res, err := queryClient.CirculatingSupply(context.Background(), req)
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryInflationRateRequest{}
			res, err := queryClient.InflationRate(context.Background(), req)
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryParamsRequest{}
			res, err := queryClient.Params(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
