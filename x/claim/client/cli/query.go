package cli

import (
	"context"
	"fmt"
	"strings"

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

	// TODO: module account balance
	claimQueryCmd.AddCommand(
		GetCmdQueryParams(),
		GetCmdQueryClaimRecord(),
		GetCmdQueryClaimableForAction(),
		GetCmdQueryTotalClaimable(),
	)

	return claimQueryCmd
}

// GetCmdQueryModuleAccountBalance implements a command to return the current balance
// of the airdrop escrow account.
func GetCmdQueryModuleAccountBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module-account",
		Short: "Query the current escrow account balance",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryModuleAccountBalanceRequest{}

			res, err := queryClient.ModuleAccountBalance(context.Background(), req)
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

// GetCmdQueryClaimRecord implements the query claim-records command.
func GetCmdQueryClaimRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-record [address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the claim record for an account.",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the claim record for an account.
This contains an address' initial claimable amounts, and the completed actions.

Example:
$ %s query claim claim-record <address>
`,
				version.AppName,
			),
		),
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

// GetCmdQueryClaimableForAction implements the query claimable for action command.
func GetCmdQueryClaimableForAction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claimable-for-action [address] [action]",
		Args:  cobra.ExactArgs(2),
		Short: "Query an address' claimable amount for a specific action",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query an address' claimable amount for a specific action

Example:
$ %s query claim claimable-for-action evmos1ey69r37gfxvxg62sh4r0ktpuc46pzjrm23kcrx ActionVote
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			action, ok := types.Action_value[args[1]]
			if !ok {
				// TODO: add actions
				return fmt.Errorf("invalid Action type: %s.  Valid actions are %s, %s", args[1],
					types.ActionVote, types.ActionDelegate)
			}

			// Query store
			res, err := queryClient.ClaimableForAction(context.Background(), &types.QueryClaimableForActionRequest{
				Address: args[0],
				Action:  types.Action(action),
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryClaimable implements the query claimables command.
func GetCmdQueryTotalClaimable() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-claimable [address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the total claimable amount remaining for an account.",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the total claimable amount remaining for an account.
Example:
$ %s query claim total-claimable evmos1ey69r37gfxvxg62sh4r0ktpuc46pzjrm23kcrx
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryTotalClaimableRequest{
				Address: args[0],
			}

			// Query store
			res, err := queryClient.TotalClaimable(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
