// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package osmosis

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/types/address"
	// "golang.org/x/exp/slices"

	"github.com/cosmos/btcutil/bech32"
	// transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
)

const (
	/// MaxSlippagePercentage is the maximum slippage percentage that can be used in the
	/// definition of the slippage for the swap.
	MaxSlippagePercentage uint8 = 20
	/// MaxWindowSeconds is the maximum number of seconds that can be used in the
	/// definition of the slippage for the swap.
	MaxWindowSeconds uint64 = 60
)

const (
	// OsmosisDenom is the base denom in the Osmosis chain.
	OsmosisDenom = "uosmo"
)

// Twap represents a Time-Weighted Average Price configuration.
type Twap struct {
	// SlippagePercentage specifies the acceptable slippage percentage for a transaction.
	SlippagePercentage uint8 `json:"slippage_percentage"`
	// WindowSeconds defines the duration for which the TWAP is calculated.
	WindowSeconds uint64 `json:"window_seconds"`
}

// Slippage specify how to compute the slippage of the swap. For this version of the outpost
// only the TWAP is allowed.
type Slippage struct {
	Twap *Twap `json:"twap"`
}

// OsmosisSwap represents the details for a swap transaction on the Osmosis chain
// using the XCS V2 contract. This payload is one of the variant of the entry_point Execute
// in the CosmWasm contract.
//
//nolint:revive
type OsmosisSwap struct {
	// OutputDenom specifies the desired output denomination for the swap.
	OutputDenom string `json:"output_denom"`
	// Twap represents the TWAP configuration for the swap.
	Slippage *Slippage `json:"slippage"`
	// Receiver is the address of the entity receiving the swapped amount.
	Receiver string `json:"receiver"`
	// OnFailedDelivery specifies the action to be taken in case the swap delivery fails.
	// This can be "do_nothing" or the address on the Osmosis chain that can recover funds
	// in case of errors.
	OnFailedDelivery string `json:"on_failed_delivery"`
	// NextMemo contains any additional memo information for the next operation in a PFM setting.
	NextMemo string `json:"next_memo,omitempty"`
}

// Msg contains the OsmosisSwap details used in the memo relayed to the Osmosis ibc hook middleware.
type Msg struct {
	// OsmosisSwap provides details for a swap transaction. It's optional and can be omitted if not provided.
	OsmosisSwap *OsmosisSwap `json:"osmosis_swap"`
}

// Memo wraps the message details for the IBC packet relyaed to the Osmosis chain. This include the
// address of the smart contract that will receive the Msg.
type Memo struct {
	// Contract represents the address or identifier of the contract to be called.
	Contract string `json:"contract"`
	// Msg contains the details of the operation to be executed on the contract.
	Msg *Msg `json:"msg"`
}

// RawPacketMetadata is the raw packet metadata used to construct a JSON string.
type RawPacketMetadata struct {
	// The Osmosis outpost IBC memo.
	Memo *Memo `json:"memo"`
}

// CreatePacketWithMemo creates the IBC packet with the memo for the Osmosis
// outpost that can be parsed by the ibc hook middleware on the Osmosis chain.
func CreatePacketWithMemo(
	outputDenom, receiver, contract string,
	slippagePercentage uint8,
	windowSeconds uint64,
	onFailedDelivery, nextMemo string,
) *RawPacketMetadata {
	return &RawPacketMetadata{
		&Memo{
			Contract: contract,
			Msg: &Msg{
				OsmosisSwap: &OsmosisSwap{
					OutputDenom: outputDenom,
					Slippage: &Slippage{
						&Twap{
							SlippagePercentage: slippagePercentage,
							WindowSeconds:      windowSeconds,
						},
					},
					Receiver:         receiver,
					OnFailedDelivery: onFailedDelivery,
					NextMemo:         nextMemo,
				},
			},
		},
	}
}

// ConvertToJSON convert the RawPacketMetadata type into a JSON formatted
// string.
func (r RawPacketMetadata) String() string {
	// Convert the struct to a JSON string
	jsonBytes, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return ""
	}

	return string(jsonBytes)
}

// Validate performs basic validation of the IBC memo for the Osmosis outpost.
// This function assumes that memo field are parsed with ParseSwapPacketData, which
// performs data casting ensuring outputDenom cannot be an empty string.
func (m Memo) Validate() error {
	osmosisSwap := m.Msg.OsmosisSwap

	if osmosisSwap.OnFailedDelivery == "" {
		return fmt.Errorf(ErrEmptyOnFailedDelivery)
	}

	// Check if account is a valid bech32 address
	_, _, err := bech32.Decode(osmosisSwap.Receiver, address.MaxAddrLen)
	if err != nil {
		return err
	}

	if osmosisSwap.Slippage.Twap.SlippagePercentage == 0 || osmosisSwap.Slippage.Twap.SlippagePercentage > MaxSlippagePercentage {
		return fmt.Errorf(ErrSlippagePercentage)
	}

	if osmosisSwap.Slippage.Twap.WindowSeconds == 0 || osmosisSwap.Slippage.Twap.WindowSeconds > MaxWindowSeconds {
		return fmt.Errorf(ErrWindowSeconds)
	}

	return nil
}

// func tmpValidate() {
// 	if osmosisSwap.OutputDenom == input {
// 		return fmt.Errorf(ErrInputEqualOutput, input)
// 	}
//
// 	osmoIBCDenom := transfertypes.DenomTrace{
// 		Path:      fmt.Sprintf("%s/%s", portID, channelID),
// 		BaseDenom: OsmosisDenom,
// 	}.IBCDenom()
//
// 	// Check that the input token is evmos or osmo.
// 	// This constraint will be removed in future
// 	validInput := []string{stakingDenom, osmoIBCDenom}
// 	if !slices.Contains(validInput, input) {
// 		return fmt.Errorf(ErrInputTokenNotSupported, validInput)
// 	}
// }

// ParseSwapPacketData parses the packet data for the Osmosis swap function.
func ParseSwapPacketData(args []interface{}) (
	sender, input, output common.Address,
	amount *big.Int,
	slippagePercentage uint8,
	windowSeconds uint64,
	receiver string,
	err error,
) {
	if len(args) != 7 {
		return common.Address{}, common.Address{}, common.Address{}, nil, 0, 0, "", fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 7, len(args))
	}

	sender, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, 0, 0, "", fmt.Errorf(cmn.ErrInvalidType, "sender", common.Address{}, args[0])
	}

	input, ok = args[1].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, 0, 0, "", fmt.Errorf(cmn.ErrInvalidType, "input", common.Address{}, args[1])
	}

	output, ok = args[2].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, 0, 0, "", fmt.Errorf(cmn.ErrInvalidType, "output", common.Address{}, args[2])
	}

	amount, ok = args[3].(*big.Int)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, 0, 0, "", fmt.Errorf(cmn.ErrInvalidType, "amount", big.Int{}, args[3])
	}

	slippagePercentage, ok = args[4].(uint8)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, 0, 0, "", fmt.Errorf(cmn.ErrInvalidType, "slippagePercentage", uint8(0), args[4])
	}

	windowSeconds, ok = args[5].(uint64)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, 0, 0, "", fmt.Errorf(cmn.ErrInvalidType, "windowSeconds", uint64(0), args[5])
	}

	receiver, ok = args[6].(string)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, 0, 0, "", fmt.Errorf(cmn.ErrInvalidType, "receiver", "", args[6])
	}

	return sender, input, output, amount, slippagePercentage, windowSeconds, receiver, nil
}
