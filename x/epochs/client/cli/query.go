package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
	"github.com/tharsis/evmos/x/epochs/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
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
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query running epoch infos.

Example:
$ %s query epochs epoch-infos
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.EpochInfos(cmd.Context(), &types.QueryEpochsInfoRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
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
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query current epoch by specified identifier.

Example:
$ %s query epochs current-epoch weekly
`,
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
