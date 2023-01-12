// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/evmos/evmos/v11/x/vesting/types"
)

// Transaction command flags
const (
	FlagDelayed  = "delayed"
	FlagDest     = "dest"
	FlagLockup   = "lockup"
	FlagMerge    = "merge"
	FlagVesting  = "vesting"
	FlagClawback = "clawback"
	FlagFunder   = "funder"
)

// NewTxCmd returns a root CLI command handler for certain modules/vesting
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
		NewMsgClawbackCmd(),
		NewMsgUpdateVestingFunderCmd(),
	)

	return txCmd
}

// NewMsgCreateClawbackVestingAccountCmd returns a CLI command handler for creating a
// MsgCreateClawbackVestingAccount transaction.
func NewMsgCreateClawbackVestingAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-clawback-vesting-account TO_ADDRESS",
		Short: "Create a new vesting account funded with an allocation of tokens, subject to clawback.",
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

			merge, _ := cmd.Flags().GetBool(FlagMerge)

			msg := types.NewMsgCreateClawbackVestingAccount(clientCtx.GetFromAddress(), toAddr, time.Unix(commonStart, 0), lockupPeriods, vestingPeriods, merge)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool(FlagMerge, false, "Merge new amount and schedule with existing ClawbackVestingAccount, if any")
	cmd.Flags().String(FlagLockup, "", "path to file containing unlocking periods")
	cmd.Flags().String(FlagVesting, "", "path to file containing vesting periods")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewMsgClawbackCmd returns a CLI command handler for creating a
// MsgClawback transaction.
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
