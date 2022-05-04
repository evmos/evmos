package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/tharsis/evmos/v4/x/fees/types"
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
		GetCmdQueryDevFeeInfos(),
		GetCmdQueryDevFeeInfo(),
		GetCmdQueryParams(),
		GetCmdQueryDevFeeInfosPerDeployer(),
	)

	return feesQueryCmd
}

// GetCmdQueryDevFeeInfos implements a command to return all registered
// contracts for fee distribution
func GetCmdQueryDevFeeInfos() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fee-infos",
		Short: "Query  all registered contracts for fee distribution",
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

			req := &types.QueryDevFeeInfosRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.DevFeeInfos(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryDevFeeInfo implements a command to return a registered contract
// for fee distribution
func GetCmdQueryDevFeeInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fee-info [address]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query a registered contract for fee distribution by hex address",
		Long:    "Query a registered contract for fee distribution by hex address",
		Example: fmt.Sprintf("%s query fees fee-info <address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryDevFeeInfoRequest{ContractAddress: args[0]}

			// Query store
			res, err := queryClient.DevFeeInfo(context.Background(), req)
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

// GetCmdQueryDevFeeInfosPerDeployer implements a command that returns all
// contracts that a deployer has registered for fee distribution
func GetCmdQueryDevFeeInfosPerDeployer() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fee-infos-deployer [address]",
		Args:    cobra.ExactArgs(1),
		Short:   "Query all contracts that a deployer has registered.",
		Long:    "Query all contracts that a deployer has registered for fee distribution.",
		Example: fmt.Sprintf("%s query fees fee-infos-deployer <address>", version.AppName),
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
			res, err := queryClient.DevFeeInfosPerDeployer(context.Background(), &types.QueryDevFeeInfosPerDeployerRequest{
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
