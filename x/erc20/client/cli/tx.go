// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/evmos/evmos/v16/x/erc20/types"
)

// NewTxCmd returns a root CLI command handler for erc20 transaction commands
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "erc20 subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	return txCmd
}

// NewRegisterERC20ProposalCmd implements the command to submit a community-pool-spend proposal
func NewRegisterERC20ProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "register-erc20 ERC20_ADDRESS...",
		Args:    cobra.MinimumNArgs(1),
		Short:   "Submit a proposal to register ERC20 token",
		Long:    "Submit a proposal to register ERC20 tokens along with an initial deposit. To register multiple tokens in one proposal pass them after each other e.g. `register-erc20 <contract-address1> <contract-address2>` ",
		Example: fmt.Sprintf("$ %s tx gov submit-legacy-proposal register-erc20 <contract-address> --from=<key_or_address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagDescription) //nolint:staticcheck
			if err != nil {
				return err
			}

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			erc20Addresses := args
			from := clientCtx.GetFromAddress()
			content := types.NewRegisterERC20Proposal(title, description, erc20Addresses...)

			msg, err := govv1beta1.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal") //nolint:staticcheck
	cmd.Flags().String(cli.FlagDeposit, "1aevmos", "deposit of proposal")
	if err := cmd.MarkFlagRequired(cli.FlagTitle); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDescription); err != nil { //nolint:staticcheck
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDeposit); err != nil {
		panic(err)
	}
	return cmd
}

// NewToggleTokenConversionProposalCmd implements the command to submit a community-pool-spend proposal
func NewToggleTokenConversionProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "toggle-token-conversion TOKEN",
		Args:    cobra.ExactArgs(1),
		Short:   "Submit a toggle token conversion proposal",
		Long:    "Submit a proposal to toggle the conversion of a token pair along with an initial deposit.",
		Example: fmt.Sprintf("$ %s tx gov submit-legacy-proposal toggle-token-conversion DENOM_OR_CONTRACT --from=<key_or_address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagDescription) //nolint:staticcheck
			if err != nil {
				return err
			}

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			token := args[0]
			content := types.NewToggleTokenConversionProposal(title, description, token)

			msg, err := govv1beta1.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal") //nolint:staticcheck
	cmd.Flags().String(cli.FlagDeposit, "1aevmos", "deposit of proposal")
	if err := cmd.MarkFlagRequired(cli.FlagTitle); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDescription); err != nil { //nolint:staticcheck
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDeposit); err != nil {
		panic(err)
	}
	return cmd
}
