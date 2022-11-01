package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/evmos/evmos/v10/x/incentives/types"
)

// GetQueryCmd returns the parent command for all incentives CLI query commands.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the incentives module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetIncentivesCmd(),
		GetIncentiveCmd(),
		GetGasMetersCmd(),
		GetGasMeterCmd(),
		GetAllocationMetersCmd(),
		GetAllocationMeterCmd(),
		GetParamsCmd(),
	)
	return cmd
}

// GetIncentivesCmd queries the list of incentives
func GetIncentivesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "incentives",
		Short: "Gets all registered incentives",
		Long:  "Gets all registered incentives",
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

			req := &types.QueryIncentivesRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.Incentives(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetIncentiveCmd queries a given contract incentive
func GetIncentiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "incentive CONTRACT_ADDRESS",
		Short: "Gets incentive for a given contract",
		Long:  "Gets incentive for a given contract",
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

			req := &types.QueryIncentiveRequest{
				Contract: args[0],
			}

			res, err := queryClient.Incentive(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetGasMetersCmd queries the list of incentives
func GetGasMetersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gas-meters CONTRACT_ADDRESS",
		Short: "Gets gas meters for a given incentive",
		Long:  "Gets gas meters for a given incentive",
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

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QueryGasMetersRequest{
				Contract:   args[0],
				Pagination: pageReq,
			}

			res, err := queryClient.GasMeters(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetGasMeterCmd queries the list of incentives
func GetGasMeterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gas-meter CONTRACT_ADDRESS PARTICIPANT_ADDRESS",
		Short: "Gets gas meter for a given incentive and user",
		Long:  "Gets gas meter for a given incentive and user",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !common.IsHexAddress(args[0]) {
				return fmt.Errorf("invalid contract address: %s", args[0])
			}

			if !common.IsHexAddress(args[1]) {
				return fmt.Errorf("invalid user address: %s", args[0])
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryGasMeterRequest{
				Contract:    args[0],
				Participant: args[1],
			}

			res, err := queryClient.GasMeter(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetAllocationMetersCmd queries the list of allocation meters
func GetAllocationMetersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "allocation-meters",
		Short: "Gets all registered allocation meters",
		Long:  "Gets all registered allocation meters",
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

			req := &types.QueryAllocationMetersRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.AllocationMeters(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetAllocationMeterCmd queries a given denom allocation meter
func GetAllocationMeterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "allocation-meter DENOM",
		Short: "Gets allocation meter for a denom",
		Long:  "Gets allocation meter for a denom",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryAllocationMeterRequest{
				Denom: args[0],
			}

			res, err := queryClient.AllocationMeter(context.Background(), req)
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
		Short: "Gets incentives params",
		Long:  "Gets incentives params",
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
