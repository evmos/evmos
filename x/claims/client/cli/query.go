package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/tharsis/evmos/v3/x/claims/types"
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
		GetCmdQueryClaimsRecords(),
		GetCmdQueryClaimsRecord(),
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

// GetCmdQueryClaimsRecords implements the query claim-records command.
func GetCmdQueryClaimsRecords() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "claims-records",
		Args:    cobra.NoArgs,
		Short:   "Query all the claims records",
		Long:    "Query the list of all the claims records",
		Example: fmt.Sprintf("%s query claims claims-records", version.AppName),
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

			req := &types.QueryClaimsRecordsRequest{
				Pagination: pageReq,
			}

			// Query store
			res, err := queryClient.ClaimsRecords(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryClaimsRecord implements the query claim-record command.
func GetCmdQueryClaimsRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "claims-record [address]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query the claims records for an account.",
		Long:    "Query the claims records for an account.\nThis contains an address' initial claimable amount, and the claims per action.",
		Example: fmt.Sprintf("%s query claims claims-record <address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			// Query store
			res, err := queryClient.ClaimsRecord(context.Background(), &types.QueryClaimsRecordRequest{Address: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
