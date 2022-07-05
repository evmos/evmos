package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/evmos/evmos/v6/x/feesplit/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	feesQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	feesQueryCmd.AddCommand(
		GetCmdQueryFeeSplits(),
		GetCmdQueryFeeSplit(),
		GetCmdQueryParams(),
		GetCmdQueryDeployerFeeSplits(),
		GetCmdQueryWithdrawerFeeSplits(),
	)

	return feesQueryCmd
}

// GetCmdQueryFeeSplits implements a command to return all registered contracts
// for fee distribution
func GetCmdQueryFeeSplits() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fee-splits",
		Short: "Query all fee splits",
		Args:  cobra.NoArgs,
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

			req := &types.QueryFeeSplitsRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.FeeSplits(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryFee implements a command to return a registered contract for fee
// distribution
func GetCmdQueryFeeSplit() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fee-split [contract-address]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query a registered contract for fee distribution by hex address",
		Long:    "Query a registered contract for fee distribution by hex address",
		Example: fmt.Sprintf("%s query feesplit fee-split <address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryFeeSplitRequest{ContractAddress: args[0]}

			// Query store
			res, err := queryClient.FeeSplit(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryParams implements a command to return the current fees
// parameters.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current fees parameters",
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

// GetCmdQueryDeployerFeeSplits implements a command that returns all contracts that
// a deployer has registered for fee distribution
func GetCmdQueryDeployerFeeSplits() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fee-splits-deployer [deployer-address]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query all fee splits that a deployer has registered",
		Long:    "Query all fee splits that a deployer has registered for fee distribution.",
		Example: fmt.Sprintf("%s query feesplit fee-splits-deployer <address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Query store
			res, err := queryClient.DeployerFeeSplits(context.Background(), &types.QueryDeployerFeeSplitsRequest{
				DeployerAddress: args[0],
				Pagination:      pageReq,
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

// GetCmdQueryWithdrawerFeeSplits implements a command that returns all fees that have
// registered for fee distribution with a given withdraw address
func GetCmdQueryWithdrawerFeeSplits() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fee-splits-withdrawer [withdrawer-address]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query all fee splits for a given withdrawer address.",
		Long:    "Query all fee splits for a given withdrawer address that have been registered for fee distribution.",
		Example: fmt.Sprintf("%s query feesplit fee-splits-withdrawer <address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Query store
			res, err := queryClient.WithdrawerFeeSplits(context.Background(), &types.QueryWithdrawerFeeSplitsRequest{
				WithdrawerAddress: args[0],
				Pagination:        pageReq,
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
