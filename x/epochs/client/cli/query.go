package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/evmos/evmos/v9/x/epochs/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group epochs queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdEpochsInfos(),
		GetCmdCurrentEpoch(),
	)

	return cmd
}

// GetCmdEpochsInfos provide running epochInfos
func GetCmdEpochsInfos() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "epoch-infos",
		Short: "Query running epochInfos",
		Example: strings.TrimSpace(
			fmt.Sprintf(`$ %s query epochs epoch-infos`,
				version.AppName,
			),
		),
		Args: cobra.NoArgs,
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

			req := &types.QueryEpochsInfoRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.EpochInfos(cmd.Context(), req)
			if err != nil {
				return err
			}

			// return clientCtx.PrintProto(res)
			return clientCtx.PrintObjectLegacy(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdCurrentEpoch provides current epoch by specified identifier
func GetCmdCurrentEpoch() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current-epoch",
		Short: "Query current epoch by specified identifier",
		Example: strings.TrimSpace(
			fmt.Sprintf(`$ %s query epochs current-epoch week`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.CurrentEpoch(cmd.Context(), &types.QueryCurrentEpochRequest{
				Identifier: args[0],
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
