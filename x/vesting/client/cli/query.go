package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/tharsis/evmos/x/vesting/types"
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
		GetUnvestedCmd(),
		GetVestedCmd(),
		GetLockedCmd(),
	)
	return cmd
}

// GetUnvestedCmd queries the unvested tokens for a given vesting account
func GetUnvestedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unvested [address]",
		Short: "Gets unvested tokens for a vesting account",
		Long:  "Gets unvested tokens for a vesting account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryUnvestedRequest{
				Address: args[0],
			}

			res, err := queryClient.Unvested(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%s\n", res.Unvested))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetVestedCmd queries the unvested tokens for a given vesting account
func GetVestedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vested [address]",
		Short: "Gets vested tokens for a vesting account",
		Long:  "Gets vested tokens for a vesting account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryVestedRequest{
				Address: args[0],
			}

			res, err := queryClient.Vested(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%s\n", res.Vested))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetLockedCmd queries the unvested tokens for a given vesting account
func GetLockedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "locked [address]",
		Short: "Gets locked tokens for a vesting account",
		Long:  "Gets locked tokens for a vesting account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryLockedRequest{
				Address: args[0],
			}

			res, err := queryClient.Locked(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%s\n", res.Locked))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
