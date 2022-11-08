package cli

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
	"github.com/CosmWasm/wasmd/x/wasm/types"
)

const (
	flagAmount                    = "amount"
	flagLabel                     = "label"
	flagAdmin                     = "admin"
	flagNoAdmin                   = "no-admin"
	flagFixMsg                    = "fix-msg"
	flagRunAs                     = "run-as"
	flagInstantiateByEverybody    = "instantiate-everybody"
	flagInstantiateNobody         = "instantiate-nobody"
	flagInstantiateByAddress      = "instantiate-only-address"
	flagInstantiateByAnyOfAddress = "instantiate-anyof-addresses"
	flagUnpinCode                 = "unpin-code"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Wasm transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
		SilenceUsage:               true,
	}
	txCmd.AddCommand(
		StoreCodeCmd(),
		InstantiateContractCmd(),
		InstantiateContract2Cmd(),
		ExecuteContractCmd(),
		MigrateContractCmd(),
		UpdateContractAdminCmd(),
		ClearContractAdminCmd(),
	)
	return txCmd
}

// StoreCodeCmd will upload code to be reused.
func StoreCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "store [wasm file]",
		Short:   "Upload a wasm binary",
		Aliases: []string{"upload", "st", "s"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg, err := parseStoreCodeArgs(args[0], clientCtx.GetFromAddress(), cmd.Flags())
			if err != nil {
				return err
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(flagInstantiateByEverybody, "", "Everybody can instantiate a contract from the code, optional")
	cmd.Flags().String(flagInstantiateNobody, "", "Nobody except the governance process can instantiate a contract from the code, optional")
	cmd.Flags().String(flagInstantiateByAddress, "", "Deprecated: Only this address can instantiate a contract from the code, optional")
	cmd.Flags().StringSlice(flagInstantiateByAnyOfAddress, []string{}, "Any of the addresses can instantiate a contract from the code, optional")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func parseStoreCodeArgs(file string, sender sdk.AccAddress, flags *flag.FlagSet) (types.MsgStoreCode, error) {
	wasm, err := os.ReadFile(file)
	if err != nil {
		return types.MsgStoreCode{}, err
	}

	// gzip the wasm file
	if ioutils.IsWasm(wasm) {
		wasm, err = ioutils.GzipIt(wasm)

		if err != nil {
			return types.MsgStoreCode{}, err
		}
	} else if !ioutils.IsGzip(wasm) {
		return types.MsgStoreCode{}, fmt.Errorf("invalid input file. Use wasm binary or gzip")
	}

	perm, err := parseAccessConfigFlags(flags)
	if err != nil {
		return types.MsgStoreCode{}, err
	}

	msg := types.MsgStoreCode{
		Sender:                sender.String(),
		WASMByteCode:          wasm,
		InstantiatePermission: perm,
	}
	return msg, nil
}

func parseAccessConfigFlags(flags *flag.FlagSet) (*types.AccessConfig, error) {
	addrs, err := flags.GetStringSlice(flagInstantiateByAnyOfAddress)
	if err != nil {
		return nil, fmt.Errorf("flag any of: %s", err)
	}
	if len(addrs) != 0 {
		acceptedAddrs := make([]sdk.AccAddress, len(addrs))
		for i, v := range addrs {
			acceptedAddrs[i], err = sdk.AccAddressFromBech32(v)
			if err != nil {
				return nil, fmt.Errorf("parse %q: %w", v, err)
			}
		}
		x := types.AccessTypeAnyOfAddresses.With(acceptedAddrs...)
		return &x, nil
	}

	onlyAddrStr, err := flags.GetString(flagInstantiateByAddress)
	if err != nil {
		return nil, fmt.Errorf("instantiate by address: %s", err)
	}
	if onlyAddrStr != "" {
		allowedAddr, err := sdk.AccAddressFromBech32(onlyAddrStr)
		if err != nil {
			return nil, sdkerrors.Wrap(err, flagInstantiateByAddress)
		}
		x := types.AccessTypeOnlyAddress.With(allowedAddr)
		return &x, nil
	}
	everybodyStr, err := flags.GetString(flagInstantiateByEverybody)
	if err != nil {
		return nil, fmt.Errorf("instantiate by everybody: %s", err)
	}
	if everybodyStr != "" {
		ok, err := strconv.ParseBool(everybodyStr)
		if err != nil {
			return nil, fmt.Errorf("boolean value expected for instantiate by everybody: %s", err)
		}
		if ok {
			return &types.AllowEverybody, nil
		}
	}

	nobodyStr, err := flags.GetString(flagInstantiateNobody)
	if err != nil {
		return nil, fmt.Errorf("instantiate by nobody: %s", err)
	}
	if nobodyStr != "" {
		ok, err := strconv.ParseBool(nobodyStr)
		if err != nil {
			return nil, fmt.Errorf("boolean value expected for instantiate by nobody: %s", err)
		}
		if ok {
			return &types.AllowNobody, nil
		}
	}
	return nil, nil
}

// InstantiateContractCmd will instantiate a contract from previously uploaded code.
func InstantiateContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instantiate [code_id_int64] [json_encoded_init_args] --label [text] --admin [address,optional] --amount [coins,optional] ",
		Short: "Instantiate a wasm contract",
		Long: fmt.Sprintf(`Creates a new instance of an uploaded wasm code with the given 'constructor' message.
Each contract instance has a unique address assigned.
Example:
$ %s tx wasm instantiate 1 '{"foo":"bar"}' --admin="$(%s keys show mykey -a)" \
  --from mykey --amount="100ustake" --label "local0.1.0" 
`, version.AppName, version.AppName),
		Aliases: []string{"start", "init", "inst", "i"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg, err := parseInstantiateArgs(args[0], args[1], clientCtx.GetFromAddress(), cmd.Flags())
			if err != nil {
				return err
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(flagAmount, "", "Coins to send to the contract during instantiation")
	cmd.Flags().String(flagLabel, "", "A human-readable name for this contract in lists")
	cmd.Flags().String(flagAdmin, "", "Address of an admin")
	cmd.Flags().Bool(flagNoAdmin, false, "You must set this explicitly if you don't want an admin")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// InstantiateContract2Cmd will instantiate a contract from previously uploaded code with predicable address generated
func InstantiateContract2Cmd() *cobra.Command {
	decoder := newArgDecoder(hex.DecodeString)
	cmd := &cobra.Command{
		Use: "instantiate2 [code_id_int64] [json_encoded_init_args] [salt] --label [text] --admin [address,optional] --amount [coins,optional] " +
			"--fix-msg [bool,optional]",
		Short: "Instantiate a wasm contract with predictable address",
		Long: fmt.Sprintf(`Creates a new instance of an uploaded wasm code with the given 'constructor' message.
Each contract instance has a unique address assigned. They are assigned automatically but in order to have predictable addresses 
for special use cases, the given 'salt' argument and '--fix-msg' parameters can be used to generate a custom address.

Predictable address example (also see '%s query wasm build-address -h'):
$ %s tx wasm instantiate2 1 '{"foo":"bar"}' $(echo -n "testing" | xxd -ps) --admin="$(%s keys show mykey -a)" \
  --from mykey --amount="100ustake" --label "local0.1.0" \
   --fix-msg 
`, version.AppName, version.AppName, version.AppName),
		Aliases: []string{"start", "init", "inst", "i"},
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			salt, err := decoder.DecodeString(args[2])
			if err != nil {
				return fmt.Errorf("salt: %w", err)
			}
			fixMsg, err := cmd.Flags().GetBool(flagFixMsg)
			if err != nil {
				return fmt.Errorf("fix msg: %w", err)
			}
			data, err := parseInstantiateArgs(args[0], args[1], clientCtx.GetFromAddress(), cmd.Flags())
			if err != nil {
				return err
			}
			msg := &types.MsgInstantiateContract2{
				Sender: data.Sender,
				Admin:  data.Admin,
				CodeID: data.CodeID,
				Label:  data.Label,
				Msg:    data.Msg,
				Funds:  data.Funds,
				Salt:   salt,
				FixMsg: fixMsg,
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(flagAmount, "", "Coins to send to the contract during instantiation")
	cmd.Flags().String(flagLabel, "", "A human-readable name for this contract in lists")
	cmd.Flags().String(flagAdmin, "", "Address of an admin")
	cmd.Flags().Bool(flagNoAdmin, false, "You must set this explicitly if you don't want an admin")
	cmd.Flags().Bool(flagFixMsg, false, "An optional flag to include the json_encoded_init_args for the predictable address generation mode")
	decoder.RegisterFlags(cmd.PersistentFlags(), "salt")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func parseInstantiateArgs(rawCodeID, initMsg string, sender sdk.AccAddress, flags *flag.FlagSet) (*types.MsgInstantiateContract, error) {
	// get the id of the code to instantiate
	codeID, err := strconv.ParseUint(rawCodeID, 10, 64)
	if err != nil {
		return nil, err
	}

	amountStr, err := flags.GetString(flagAmount)
	if err != nil {
		return nil, fmt.Errorf("amount: %s", err)
	}
	amount, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return nil, fmt.Errorf("amount: %s", err)
	}
	label, err := flags.GetString(flagLabel)
	if err != nil {
		return nil, fmt.Errorf("label: %s", err)
	}
	if label == "" {
		return nil, errors.New("label is required on all contracts")
	}
	adminStr, err := flags.GetString(flagAdmin)
	if err != nil {
		return nil, fmt.Errorf("admin: %s", err)
	}
	noAdmin, err := flags.GetBool(flagNoAdmin)
	if err != nil {
		return nil, fmt.Errorf("no-admin: %s", err)
	}

	// ensure sensible admin is set (or explicitly immutable)
	if adminStr == "" && !noAdmin {
		return nil, fmt.Errorf("you must set an admin or explicitly pass --no-admin to make it immutible (wasmd issue #719)")
	}
	if adminStr != "" && noAdmin {
		return nil, fmt.Errorf("you set an admin and passed --no-admin, those cannot both be true")
	}

	// build and sign the transaction, then broadcast to Tendermint
	msg := types.MsgInstantiateContract{
		Sender: sender.String(),
		CodeID: codeID,
		Label:  label,
		Funds:  amount,
		Msg:    []byte(initMsg),
		Admin:  adminStr,
	}
	return &msg, nil
}

// ExecuteContractCmd will instantiate a contract from previously uploaded code.
func ExecuteContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "execute [contract_addr_bech32] [json_encoded_send_args] --amount [coins,optional]",
		Short:   "Execute a command on a wasm contract",
		Aliases: []string{"run", "call", "exec", "ex", "e"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg, err := parseExecuteArgs(args[0], args[1], clientCtx.GetFromAddress(), cmd.Flags())
			if err != nil {
				return err
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(flagAmount, "", "Coins to send to the contract along with command")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func parseExecuteArgs(contractAddr string, execMsg string, sender sdk.AccAddress, flags *flag.FlagSet) (types.MsgExecuteContract, error) {
	amountStr, err := flags.GetString(flagAmount)
	if err != nil {
		return types.MsgExecuteContract{}, fmt.Errorf("amount: %s", err)
	}

	amount, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return types.MsgExecuteContract{}, err
	}

	return types.MsgExecuteContract{
		Sender:   sender.String(),
		Contract: contractAddr,
		Funds:    amount,
		Msg:      []byte(execMsg),
	}, nil
}
