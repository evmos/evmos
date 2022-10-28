package cli

import (
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/evmos/evmos/v9/x/incentives/types"
)

// NewRegisterIncentiveProposalCmd implements the command to submit a register
//
//	incentive proposal
//
//nolint:staticcheck // we use deprecated flags
func NewRegisterIncentiveProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "register-incentive [contract-address] [allocation] [epochs]",
		Args:    cobra.ExactArgs(3),
		Short:   "Submit a proposal to register a contract incentive",
		Long:    "Submit a proposal to register a contract incentive.",
		Example: fmt.Sprintf("$ %s tx gov submit-proposal register-incentive <contract> --from=<key_or_address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagDescription)
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

			if !common.IsHexAddress(args[0]) {
				return fmt.Errorf("invalid contract address: %s", args[0])
			}

			allocation, err := sdk.ParseDecCoins(args[1])
			if err != nil {
				return err
			}

			epochs, err := strconv.ParseUint(args[2], 10, 32)
			if err != nil {
				return err
			}

			contract := args[0]

			from := clientCtx.GetFromAddress()
			content := types.NewRegisterIncentiveProposal(title, description, contract, allocation, uint32(epochs))

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
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "1aevmos", "deposit of proposal")
	if err := cmd.MarkFlagRequired(cli.FlagTitle); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDescription); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDeposit); err != nil {
		panic(err)
	}
	return cmd
}

// NewCancelIncentiveProposalCmd implements the command to submit a cancel
//
//	incentive proposal
//
//nolint:staticcheck
func NewCancelIncentiveProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cancel-incentive [contract-address]",
		Args:    cobra.ExactArgs(1),
		Short:   "Submit a proposal to cancel a contract incentive",
		Long:    "Submit a proposal to cancel a contract incentive.",
		Example: fmt.Sprintf("$ %s tx gov submit-proposal cancel-incentive <contract> --from=<key_or_address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagDescription)
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

			if !common.IsHexAddress(args[0]) {
				return fmt.Errorf("invalid contract address: %s", args[0])
			}

			contract := args[0]

			from := clientCtx.GetFromAddress()
			content := types.NewCancelIncentiveProposal(title, description, contract)

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
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "1aevmos", "deposit of proposal")
	if err := cmd.MarkFlagRequired(cli.FlagTitle); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDescription); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(cli.FlagDeposit); err != nil {
		panic(err)
	}
	return cmd
}
