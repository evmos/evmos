// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/evmos/evmos/v19/x/epochs/types"
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
		GetCmdStateRentInfos(),
	)

	return cmd
}

// GetCmdStateRentInfos provide running epochInfos
func GetCmdStateRentInfos() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "",
		Example: strings.TrimSpace(
			fmt.Sprintf(`$ %s query staterent info`,
				version.AppName,
			),
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Println("hello hello")
			// clientCtx, err := client.GetClientQueryContext(cmd)
			// if err != nil {
			// 	return err
			// }
			// queryClient := types.NewQueryClient(clientCtx)
			//
			// pageReq, err := client.ReadPageRequest(cmd.Flags())
			// if err != nil {
			// 	return err
			// }
			//
			// req := &types.QueryEpochsInfoRequest{
			// 	Pagination: pageReq,
			// }
			//
			// res, err := queryClient.EpochInfos(cmd.Context(), req)
			// if err != nil {
			// 	return err
			// }
			//
			// // return clientCtx.PrintProto(res)
			// return clientCtx.PrintObjectLegacy(res)
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
