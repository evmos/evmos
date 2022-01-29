package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/tharsis/evmos/x/inflation/types"
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
		Use:   "epoch--mint-provisions",
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
