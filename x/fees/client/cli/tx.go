package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	ethermint "github.com/tharsis/ethermint/types"

	"github.com/tharsis/evmos/v3/x/fees/types"
)

// NewTxCmd returns a root CLI command handler for certain modules/erc20 transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "fees subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewRegisterDevFeeInfo(),
		NewCancelDevFeeInfo(),
		NewUpdateDevFeeInfo(),
	)
	return txCmd
}

// NewRegisterDevFeeInfo returns a CLI command handler for registering a
// contract for fee distribution
func NewRegisterDevFeeInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-fee [contract_hex] [withdraw_bech32]",
		Short: "Register a contract for fee distribution",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var withdraw string
			deployer := cliCtx.GetFromAddress()

			contract := args[0]
			if err := ethermint.ValidateAddress(contract); err != nil {
				return fmt.Errorf("invalid contract hex address %w", err)
			}

			if len(args) == 2 {
				withdraw = args[1]
				if _, err := sdk.AccAddressFromBech32(withdraw); err != nil {
					return fmt.Errorf("invalid withdraw bech32 address %w", err)
				}
			}

			msg := &types.MsgRegisterDevFeeInfo{
				ContractAddress: contract,
				DeployerAddress: deployer.String(),
				WithdrawAddress: withdraw,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewCancelDevFeeInfo returns a CLI command handler for canceling a
// contract for fee distribution
func NewCancelDevFeeInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-fee [contract_hex]",
		Short: "Cancel a contract from fee distribution",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			deployer := cliCtx.GetFromAddress()

			contract := args[0]
			if err := ethermint.ValidateAddress(contract); err != nil {
				return fmt.Errorf("invalid contract hex address %w", err)
			}

			msg := &types.MsgCancelDevFeeInfo{
				ContractAddress: contract,
				DeployerAddress: deployer.String(),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewUpdateDevFeeInfo returns a CLI command handler for updating the withdraw
// address of a contract for fee distribution
func NewUpdateDevFeeInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-fee [contract_hex] [withdraw_bech32]",
		Short: "Update withdraw address for a contract registered for fee distribution",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			deployer := cliCtx.GetFromAddress()

			contract := args[0]
			if err := ethermint.ValidateAddress(contract); err != nil {
				return fmt.Errorf("invalid contract hex address %w", err)
			}

			withdraw := args[1]
			if _, err := sdk.AccAddressFromBech32(withdraw); err != nil {
				return fmt.Errorf("invalid withdraw bech32 address %w", err)
			}

			msg := &types.MsgUpdateDevFeeInfo{
				ContractAddress: contract,
				DeployerAddress: deployer.String(),
				WithdrawAddress: withdraw,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
