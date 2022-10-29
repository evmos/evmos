package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	ethermint "github.com/evmos/ethermint/types"

	"github.com/evmos/evmos/v10/x/revenue/types"
)

// NewTxCmd returns a root CLI command handler for certain modules/revenue
// transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "revenue subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewRegisterRevenue(),
		NewCancelRevenue(),
		NewUpdateRevenue(),
	)
	return txCmd
}

// NewRegisterRevenue returns a CLI command handler for registering a
// contract for fee distribution
func NewRegisterRevenue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register [contract_hex] [nonces] [withdraw_bech32]",
		Short: "Register a contract for fee distribution. **NOTE** Please ensure, that the deployer of the contract (or the factory that deployes the contract) is an account that is owned by your project, to avoid that an individual deployer who leaves your project becomes malicious.",
		Long:  "Register a contract for fee distribution.\nOnly the contract deployer can register a contract.\nProvide the account nonce(s) used to derive the contract address. E.g.: you have an account nonce of 4 when you send a deployment transaction for a contract A; you use this contract as a factory, to create another contract B. If you register A, the nonces value is \"4\". If you register B, the nonces value is \"4,1\" (B is the first contract created by A). \nThe withdraw address defaults to the deployer address if not provided.",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var withdraw string
			deployer := cliCtx.GetFromAddress()

			contract := args[0]
			if err := ethermint.ValidateNonZeroAddress(contract); err != nil {
				return fmt.Errorf("invalid contract hex address %w", err)
			}

			var nonces []uint64
			if err = json.Unmarshal([]byte("["+args[1]+"]"), &nonces); err != nil {
				return fmt.Errorf("invalid nonces %w", err)
			}

			if len(args) == 3 {
				withdraw = args[2]
				if _, err := sdk.AccAddressFromBech32(withdraw); err != nil {
					return fmt.Errorf("invalid withdraw bech32 address %w", err)
				}
			}

			// If withdraw address is the same as contract deployer, remove the
			// field for avoiding storage bloat
			if deployer.String() == withdraw {
				withdraw = ""
			}

			msg := &types.MsgRegisterRevenue{
				ContractAddress:   contract,
				DeployerAddress:   deployer.String(),
				WithdrawerAddress: withdraw,
				Nonces:            nonces,
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

// NewCancelRevenue returns a CLI command handler for canceling a
// contract for fee distribution
func NewCancelRevenue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel [contract_hex]",
		Short: "Cancel a contract from fee distribution",
		Long:  "Cancel a contract from fee distribution. The deployer will no longer receive fees from users interacting with the contract. \nOnly the contract deployer can cancel a contract.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			deployer := cliCtx.GetFromAddress()

			contract := args[0]
			if err := ethermint.ValidateNonZeroAddress(contract); err != nil {
				return fmt.Errorf("invalid contract hex address %w", err)
			}

			msg := &types.MsgCancelRevenue{
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

// NewUpdateRevenue returns a CLI command handler for updating the withdraw
// address of a contract for fee distribution
func NewUpdateRevenue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [contract_hex] [withdraw_bech32]",
		Short: "Update withdraw address for a contract registered for fee distribution.",
		Long:  "Update withdraw address for a contract registered for fee distribution. \nOnly the contract deployer can update the withdraw address.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			deployer := cliCtx.GetFromAddress()

			contract := args[0]
			if err := ethermint.ValidateNonZeroAddress(contract); err != nil {
				return fmt.Errorf("invalid contract hex address %w", err)
			}

			withdraw := args[1]
			if _, err := sdk.AccAddressFromBech32(withdraw); err != nil {
				return fmt.Errorf("invalid withdraw bech32 address %w", err)
			}

			msg := &types.MsgUpdateRevenue{
				ContractAddress:   contract,
				DeployerAddress:   deployer.String(),
				WithdrawerAddress: withdraw,
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
