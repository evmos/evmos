package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/tharsis/evmos/x/claim/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	claimQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	claimQueryCmd.AddCommand(
		GetCmdQueryTotalUnclaimed(),
		GetCmdQueryParams(),
		GetCmdQueryClaimRecords(),
		GetCmdQueryClaimRecord(),
	)

	return claimQueryCmd
}

// GetCmdQueryTotalUnclaimed implements a command to return the current balance
// of the airdrop escrow account.
func GetCmdQueryTotalUnclaimed() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-unclaimed",
		Short: "Query the total amount of unclaimed tokens from the airdrop",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryTotalUnclaimedRequest{}

			res, err := queryClient.TotalUnclaimed(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryParams implements a command to return the current claim
// parameters.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current claims parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
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

// GetCmdQueryClaimRecords implements the query claim-records command.
func GetCmdQueryClaimRecords() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "claim-records",
		Args:    cobra.NoArgs,
		Short:   "Query all the claim records",
		Long:    "Query the list of all the claim records",
		Example: fmt.Sprintf("%s query claim claim-records", version.AppName),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QueryClaimRecordsRequest{
				Pagination: pageReq,
			}

			// Query store
			res, err := queryClient.ClaimRecords(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryClaimRecord implements the query claim-record command.
func GetCmdQueryClaimRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "claim-record [address]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query the claim records for an account.",
		Long:    "Query the claim records for an account.\nThis contains an address' initial claimable amount, and the claims per action.",
		Example: fmt.Sprintf("%s query claim claim-record <address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			// Query store
			res, err := queryClient.ClaimRecord(context.Background(), &types.QueryClaimRecordRequest{Address: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
