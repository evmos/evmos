// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/evmos/evmos/v16/x/vesting/types"
)

// Transaction command flags
const (
	FlagDest     = "dest"
	FlagLockup   = "lockup"
	FlagVesting  = "vesting"
	FlagClawback = "clawback"
	FlagFunder   = "funder"
)

// NewTxCmd returns a root CLI command handler for vesting
// transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Vesting transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewMsgCreateClawbackVestingAccountCmd(),
		NewMsgFundVestingAccountCmd(),
		NewMsgClawbackCmd(),
		NewMsgUpdateVestingFunderCmd(),
		NewMsgConvertVestingAccountCmd(),
	)

	return txCmd
}

// NewMsgCreateClawbackVestingAccountCmd returns a CLI command handler for creating a
// clawback vesting account.
func NewMsgCreateClawbackVestingAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-clawback-vesting-account FUNDER_ADDRESS ENABLE_GOV_CLAWBACK",
		Short: "Create a new vesting account at the address of the sender with a designated funder.",
		Long: `A new clawback vesting account is created for the sender account, if it is not already of such type.
Only the designated funder will be able to define lockup and vesting schedules and has to do so
using the fund-vesting-account subcommand. Clawback via governance is enabled through the second argument.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			funder, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			enableGovClawback, err := strconv.ParseBool(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateClawbackVestingAccount(funder, clientCtx.GetFromAddress(), enableGovClawback)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewMsgFundVestingAccountCmd returns a CLI command handler for funding a clawback vesting account.
func NewMsgFundVestingAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fund-vesting-account TO_ADDRESS",
		Short: "Fund a vesting account with an allocation of tokens.",
		Long: `Must provide a lockup periods file (--lockup), a vesting periods file (--vesting), or both.
If both files are given, they must describe schedules for the same total amount.
If one file is omitted, it will default to a schedule that immediately unlocks or vests the entire amount.
The described amount of coins will be transferred from the --from address to the vesting account.
Unvested coins may be "clawed back" by the funder with the clawback command.
Coins may not be transferred out of the account if they are locked or unvested. Only vested coins may be staked.

A periods file is a JSON object describing a sequence of unlocking or vesting events,
with a start time and an array of coins strings and durations relative to the start or previous event.`,
		Example: `Sample period file contents:
{
  "start_time": 1625204910,
  "periods": [
    {
      "coins": "10test",
      "length_seconds": 2592000 //30 days
    },
    {
      "coins": "10test",
      "length_seconds": 2592000 //30 days
    }
  ]
}`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				lockupStart, vestingStart     int64
				lockupPeriods, vestingPeriods sdkvesting.Periods
			)

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			toAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			lockupFile, _ := cmd.Flags().GetString(FlagLockup)
			vestingFile, _ := cmd.Flags().GetString(FlagVesting)
			if lockupFile == "" && vestingFile == "" {
				return fmt.Errorf("must specify at least one of %s or %s", FlagLockup, FlagVesting)
			}
			if lockupFile != "" {
				lockupStart, lockupPeriods, err = ReadScheduleFile(lockupFile)
				if err != nil {
					return err
				}
			}
			if vestingFile != "" {
				vestingStart, vestingPeriods, err = ReadScheduleFile(vestingFile)
				if err != nil {
					return err
				}
			}

			commonStart, _ := types.AlignSchedules(lockupStart, vestingStart, lockupPeriods, vestingPeriods)

			msg := types.NewMsgFundVestingAccount(clientCtx.GetFromAddress(), toAddr, time.Unix(commonStart, 0), lockupPeriods, vestingPeriods)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagLockup, "", "path to file containing unlocking periods")
	cmd.Flags().String(FlagVesting, "", "path to file containing vesting periods")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewMsgClawbackCmd returns a CLI command handler for clawing back unvested funds.
func NewMsgClawbackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clawback ADDRESS",
		Short: "Transfer unvested amount out of a ClawbackVestingAccount.",
		Long: `Must be requested by the original funder address (--from).
		May provide a destination address (--dest), otherwise the coins return to the funder.
		Delegated or undelegating staking tokens will be transferred in the delegated (undelegating) state.
		The recipient is vulnerable to slashing, and must act to unbond the tokens if desired.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			var dest sdk.AccAddress
			destString, _ := cmd.Flags().GetString(FlagDest)
			if destString != "" {
				dest, err = sdk.AccAddressFromBech32(destString)
				if err != nil {
					return fmt.Errorf("bad dest address: %w", err)
				}
			}

			msg := types.NewMsgClawback(clientCtx.GetFromAddress(), addr, dest)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagDest, "", "address of destination (defaults to funder)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewMsgUpdateVestingFunderCmd returns a CLI command handler for updating
// the funder of a ClawbackVestingAccount.
func NewMsgUpdateVestingFunderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-vesting-funder VESTING_ACCOUNT_ADDRESS NEW_FUNDER_ADDRESS",
		Short: "Update the funder account of an existing ClawbackVestingAccount.",
		Long: `Must be requested by the original funder address (--from).
		Need to provide the target VESTING_ACCOUNT_ADDRESS to update and the NEW_FUNDER_ADDRESS.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			vestingAcc, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			newFunder, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateVestingFunder(clientCtx.GetFromAddress(), newFunder, vestingAcc)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewMsgConvertVestingAccountCmd returns a CLI command handler for converting
// a clawback vesting account into a non-vesting account.
func NewMsgConvertVestingAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert VESTING_ACCOUNT_ADDRESS",
		Short: "Convert a vesting account to the chain's default account type.",
		Long: "Convert a vesting account to the chain's default account type. " +
			"The vesting account must be of type ClawbackVestingAccount and have all of its coins vested in order to convert" +
			"it back to the chain default account type.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgConvertVestingAccount(addr)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewClawbackProposalCmd implements the command to submit
// a proposal to clawback funds from a specified vesting account,
// that has this functionality enabled.
//
//nolint:staticcheck
func NewClawbackProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clawback ADDRESS [DEST_ADDRESS]",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Submit a proposal to clawback funds from a ClawbackVestingAccount",
		Long:  "Submit a proposal to clawback the tokens from a ClawbackVestingAccount that has this functionality enabled.",
		Example: fmt.Sprintf(
			`$ %s tx gov submit-legacy-proposal clawback <address> \
--from=<key_or_address> \
--title=<proposal_title> \
--description=<proposal_description> \
--deposit=<deposit>`,
			version.AppName,
		),
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

			from := clientCtx.GetFromAddress()

			vestingAddress := args[0]

			var destinationAddr string
			if len(args) == 2 {
				destinationAddr = args[1]
			}

			// check that args[0] is valid address in ValidateBasic()
			content := types.NewClawbackProposal(title, description, vestingAddress, destinationAddr)

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
