// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package debug

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/evmos/evmos/v16/ethereum/eip712"
	evmos "github.com/evmos/evmos/v16/types"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ethereum/go-ethereum/common"

	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
)

var flagPrefix = "prefix"

// Cmd creates a main CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Tool for helping with debugging your application",
		RunE:  client.ValidateCmd,
	}

	cmd.AddCommand(PubkeyCmd())
	cmd.AddCommand(AddrCmd())
	cmd.AddCommand(RawBytesCmd())
	cmd.AddCommand(LegacyEIP712Cmd())

	return cmd
}

// getPubKeyFromString decodes SDK PubKey using JSON marshaler.
func getPubKeyFromString(ctx client.Context, pkstr string) (cryptotypes.PubKey, error) {
	var pk cryptotypes.PubKey
	err := ctx.Codec.UnmarshalInterfaceJSON([]byte(pkstr), &pk)
	return pk, err
}

func PubkeyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pubkey [pubkey]",
		Short: "Decode a pubkey from proto JSON",
		Long:  "Decode a pubkey from proto JSON and display it's address",
		Example: fmt.Sprintf(
			`"$ %s debug pubkey '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AurroA7jvfPd1AadmmOvWM2rJSwipXfRf8yD6pLbA2DJ"}'`, //gitleaks:allow
			version.AppName,
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			pk, err := getPubKeyFromString(clientCtx, args[0])
			if err != nil {
				return err
			}

			addr := pk.Address()
			cmd.Printf("Address (EIP-55): %s\n", common.BytesToAddress(addr))
			cmd.Printf("Bech32 Acc: %s\n", sdk.AccAddress(addr))
			cmd.Println("PubKey Hex:", hex.EncodeToString(pk.Bytes()))
			return nil
		},
	}
}

func AddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addr [address]",
		Short: "Convert an address between hex and bech32",
		Long:  "Convert an address between hex encoding and bech32.",
		Example: fmt.Sprintf(
			`$ %s debug addr evmos1qqqqhe5pnaq5qq39wqkn957aydnrm45sdn8583
$ %s debug addr 0x00000Be6819f41400225702D32d3dd23663Dd690 --prefix evmos`, version.AppName, version.AppName),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addrString := args[0]
			switch {
			case common.IsHexAddress(addrString):
				addr := common.HexToAddress(addrString).Bytes()
				cmd.Println("Address bytes:", addr)

				prefix, err := cmd.Flags().GetString(flagPrefix)
				if err != nil {
					return err
				}
				if prefix == "" {
					bech32AccAddress := sdk.AccAddress(addr)

					bech32ValAddress := sdk.ValAddress(addr)

					cmd.Printf("Bech32 Acc %s\n", bech32AccAddress)
					cmd.Printf("Bech32 Val %s\n", bech32ValAddress)
				} else {
					bech32Address, err := sdk.Bech32ifyAddressBytes(prefix, addr)
					if err != nil {
						return err
					}

					cmd.Printf("Bech32 %s\n", bech32Address)
				}
			default:
				prefix := strings.SplitN(addrString, "1", 2)[0]
				hexAddr, err := sdk.GetFromBech32(addrString, prefix)
				if err != nil {
					return err
				}

				hexAddrString := common.BytesToAddress(hexAddr).String()

				cmd.Println("Address bytes:", hexAddr)
				cmd.Printf("Address hex: %s\n", hexAddrString)
			}
			return nil
		},
	}

	cmd.Flags().String(flagPrefix, "", "Bech32 encoded account prefix, for example evmos, evmosvaloper")
	return cmd
}

func RawBytesCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "raw-bytes [raw-bytes]",
		Short:   "Convert raw bytes output (eg. [10 21 13 255]) to hex",
		Example: fmt.Sprintf(`$ %s debug raw-bytes [72 101 108 108 111 44 32 112 108 97 121 103 114 111 117 110 100]`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			stringBytes := args[0]
			stringBytes = strings.Trim(stringBytes, "[")
			stringBytes = strings.Trim(stringBytes, "]")
			spl := strings.Split(stringBytes, " ")

			byteArray := []byte{}
			for _, s := range spl {
				b, err := strconv.ParseInt(s, 10, 8)
				if err != nil {
					return err
				}
				byteArray = append(byteArray, byte(b))
			}
			fmt.Printf("%X\n", byteArray)
			return nil
		},
	}
}

// LegacyEIP712Cmd outputs types of legacy EIP712 typed data
func LegacyEIP712Cmd() *cobra.Command {
	return &cobra.Command{
		Use:     "legacy-eip712 [file]",
		Short:   "Output types of legacy eip712 typed data according to the given transaction",
		Example: fmt.Sprintf(`$ %s debug legacy-eip712 tx.json --chain-id evmosd_9000-1`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			stdTx, err := authclient.ReadTxFromFile(clientCtx, args[0])
			if err != nil {
				return errors.Wrap(err, "read tx from file")
			}

			txBytes, err := clientCtx.TxConfig.TxJSONEncoder()(stdTx)
			if err != nil {
				return errors.Wrap(err, "encode tx")
			}

			chainID, err := evmos.ParseChainID(clientCtx.ChainID)
			if err != nil {
				return errors.Wrap(err, "invalid chain ID passed as argument")
			}

			td, err := eip712.LegacyWrapTxToTypedData(clientCtx.Codec, chainID.Uint64(), stdTx.GetMsgs()[0], txBytes, nil)
			if err != nil {
				return errors.Wrap(err, "wrap tx to typed data")
			}

			bz, err := json.Marshal(td.Map()["types"])
			if err != nil {
				return err
			}

			fmt.Println(string(bz))
			return nil
		},
	}
}
