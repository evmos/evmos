package cli

import (
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/tharsis/evmos/x/fees/types"
)

// NewTxCmd returns the parent command for all fee distribution CLI query commands.
func NewTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Transaction commands for the fee distribution module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewRegisterContractCmd(),
	)
	return cmd
}

// NewRegisterContractCmd implements the command to submit a register
//  incentive proposal
func NewRegisterContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "register-contract [contract-address] [nonce]",
		Args:    cobra.ExactArgs(2),
		Short:   "Register a contract to receive tx fees",
		Long:    "Register a contract te receive tx fees. The signer (from) must be de contract deployer/owner",
		Example: fmt.Sprintf("$ %s tx fees register-contract <contract> <nonce> --from=<key_or_address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if !common.IsHexAddress(args[0]) {
				return fmt.Errorf("invalid contract address: %s", args[0])
			}

			nonce, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			contract := common.HexToAddress(args[0])
			deployerAddr := common.BytesToAddress(clientCtx.GetFromAddress())

			msg := types.NewMsgRegisterContract(contract, deployerAddr, nonce)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}

// NewUpdateWithdrawAddressCmd implements the command to update the withdraw address
// of a given contract
func NewUpdateWithdrawAddressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-withdraw-address [contract-address] [new-withdraw-address]",
		Args:    cobra.ExactArgs(2),
		Short:   "Update the withdraw address of a given contract",
		Long:    "Update the withdraw address of a given contract. The signer (from) must be the current withdraw address",
		Example: fmt.Sprintf("$ %s tx fees update-withdraw-address <contract> <new-withdraw-address> --from=<key_or_address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if !common.IsHexAddress(args[0]) {
				return fmt.Errorf("invalid contract address: %s", args[0])
			}

			if !common.IsHexAddress(args[1]) {
				return fmt.Errorf("invalid contract address: %s", args[1])
			}

			contract := common.HexToAddress(args[0])
			newWithdrawAddress := common.HexToAddress(args[1])
			withdrawAddress := common.BytesToAddress(clientCtx.GetFromAddress())

			msg := types.NewMsgUpdateWithdawAddress(contract, withdrawAddress, newWithdrawAddress)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}
