package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"github.com/tharsis/evmos/x/fees/types"
)

// GetQueryCmd returns the parent command for all fee distribution CLI query commands.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the fee distribution module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetWithdrawAddressesCmd(),
		GetWithdrawAddressCmd(),
		GetParamsCmd(),
	)
	return cmd
}

// GetWithdrawAddressesCmd queries the list of registered contracts
func GetWithdrawAddressesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-addresses",
		Short: "Gets all registered fee withdraw addresses.",
		Long:  "Gets all registered dApps (contracts) with their corresponding fee withdraw addresses.",
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

			req := &types.QueryWithdrawAddressesRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.WithdrawAddresses(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetWithdrawAddressCmd queries a given contract incentive
func GetWithdrawAddressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-address [contract-address]",
		Short: "Gets the withdraw address for a registered dApp (contract)",
		Long:  "Gets the withdraw address for a registered dApp (contract)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !common.IsHexAddress(args[0]) {
				return fmt.Errorf("invalid contract address: %s", args[0])
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryWithdrawAddressRequest{
				ContractAddress: args[0],
			}

			res, err := queryClient.WithdrawAddress(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetParamsCmd queries the module parameters
func GetParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Gets fee distribution params",
		Long:  "Gets fee distribution params",
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

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
