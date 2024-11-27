// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"

	evmostypes "github.com/evmos/evmos/v20/types"
	"github.com/evmos/evmos/v20/x/erc20/types"
)

const (
	FlagAddress = "address"
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

	txCmd.AddCommand(
		NewConvertERC20Cmd(),
		NewMintCmd(),
		NewBurnCmd(),
	)

	return txCmd
}

// NewConvertERC20Cmd returns a CLI command handler for converting an ERC20
func NewConvertERC20Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert-erc20 CONTRACT_ADDRESS AMOUNT [RECEIVER]",
		Short: "Convert an ERC20 token to Cosmos coin.  When the receiver [optional] is omitted, the Cosmos coins are transferred to the sender.",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			contract := args[0]
			if err := evmostypes.ValidateAddress(contract); err != nil {
				return fmt.Errorf("invalid ERC20 contract address %w", err)
			}

			amount, ok := math.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid amount %s", args[1])
			}

			from := common.BytesToAddress(cliCtx.GetFromAddress().Bytes())

			receiver := cliCtx.GetFromAddress()
			if len(args) == 3 {
				receiver, err = sdk.AccAddressFromBech32(args[2])
				if err != nil {
					return err
				}
			}

			msg := &types.MsgConvertERC20{
				ContractAddress: contract,
				Amount:          amount,
				Receiver:        receiver.String(),
				Sender:          from.Hex(),
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewMintCmd implements the command to mint an ERC20 token
func NewMintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint CONTRACT_ADDRESS TO AMOUNT",
		Args:  cobra.ExactArgs(3),
		Short: "Mint an ERC20 token",
		Long:  "Mint an ERC20 token",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			contract := args[0]
			if err := evmostypes.ValidateAddress(contract); err != nil {
				return fmt.Errorf("invalid ERC20 contract address %w", err)
			}

			to, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			amount, ok := math.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("invalid amount %s", args[2])
			}

			from := clientCtx.GetFromAddress()

			msg := &types.MsgMint{
				ContractAddress: contract,
				Amount:          amount,
				Sender:          from.String(),
				To:              to.String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewBurnCmd implements the command to burn an ERC20 token
func NewBurnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn CONTRACT_ADDRESS AMOUNT",
		Args:  cobra.ExactArgs(2),
		Short: "Burn an ERC20 token",
		Long:  "Burn an ERC20 token",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			contract := args[0]
			if err := evmostypes.ValidateAddress(contract); err != nil {
				return fmt.Errorf("invalid ERC20 contract address %w", err)
			}

			amount, ok := math.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid amount %s", args[1])
			}

			msg := &types.MsgBurn{
				ContractAddress: contract,
				Amount:          amount,
				Sender:          clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
