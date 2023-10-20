package osmosis

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v14/ibc"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
)

// Twap represents a Time-Weighted Average Price configuration.
type Twap struct {
	// SlippagePercentage specifies the acceptable slippage percentage for a transaction.
	SlippagePercentage string `json:"slippage_percentage"`
	// WindowSeconds defines the duration for which the TWAP is calculated.
	WindowSeconds int64 `json:"window_seconds"`
}

// Slippage specify how to compute the slippage of the swap. For this version of the outpost
// only the TWAP is allowed.
type Slippage struct {
	Twap Twap `json:"twap"`
}

// OsmosisSwap represents the details for a swap transaction on the Osmosis chain
// using the XCS V2 contract. This payload is one of the variant of the entry_point Execute
// in the CosmWasm contract.
type OsmosisSwap struct {
	// OutputDenom specifies the desired output denomination for the swap.
	OutputDenom string `json:"output_denom"`
	// Twap represents the TWAP configuration for the swap.
	Slippage Slippage `json:"slippage"`
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
	Msg Msg `json:"msg"`
}

// RawPacketMetadata is the raw packet metadata used to construct a JSON string
type RawPacketMetadata struct {
	// The Osmosis outpost IBC memo. 
	Memo Memo `json:"memo"`
}

// CreateMemo creates the IBC memo for the Osmosis outpost that can be parsed by the ibc hook
// middleware on the Osmosis chain.
func CreateMemo(
	outputDenom, receiver, contract, slippagePercentage string, 
	windowSeconds int64, 
	onFailedDelivery string,
) (string, error) {

	data := &RawPacketMetadata{
		Memo{
			Contract: contract,
			Msg: Msg{
				OsmosisSwap: &OsmosisSwap{
					OutputDenom: outputDenom,
					Slippage: Slippage{
						Twap{
							SlippagePercentage: slippagePercentage,
							WindowSeconds:      windowSeconds,
						},
					},
					Receiver:         receiver,
					OnFailedDelivery: onFailedDelivery,
					// NextMemo:         "",
				},
			},
		},
	}

	// Convert the struct to a JSON string
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Print the JSON string
	fmt.Println(string(jsonBytes))
	return string(jsonBytes), nil
}

// parseSwapPacketData parses the packet data for the Osmosis swap function.
func ParseSwapPacketData(args []interface{}) (
	sender, input, output common.Address, 
	amount *big.Int,
	slippagePercentage string,
	windowSeconds uint64,
	receiver string,
	err error,
) {
	if len(args) != 7 {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", 0, "", fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 5, len(args))
	}

	sender, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", 0, "", fmt.Errorf(cmn.ErrInvalidType,"sender", common.Address{}, args[0])
	}

	input, ok = args[1].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", 0, "", fmt.Errorf(cmn.ErrInvalidType,"sender", common.Address{}, args[1])
	}

	output, ok = args[2].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", 0, "", fmt.Errorf(cmn.ErrInvalidType, "output", common.Address{}, args[2])
	}

	amount, ok = args[3].(*big.Int)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", 0, "", fmt.Errorf(cmn.ErrInvalidType,"amount", big.Int{}, args[3])
	}

	slippagePercentage, ok = args[4].(string)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", 0, "", fmt.Errorf(cmn.ErrInvalidType, "slippagePercentage", "", args[4])
	}

	windowSeconds, ok = args[5].(uint64)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", 0, "", fmt.Errorf(cmn.ErrInvalidType, "windowSeconds", uint64(0), args[4])
	}

	receiver, ok = args[6].(string)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", 0, "", fmt.Errorf(cmn.ErrInvalidType, "receiver", "", args[4])
	}

	return sender, input, output, amount, slippagePercentage,  windowSeconds, receiver, nil
}

// validateSwap performs validation on input and output denom.
func ValidateSwap(
	ctx sdk.Context,
	portID,
	channelID,
	input,
	output,
	stakingDenom string,
) (err error) {
	// input and output cannot be equal
	if input == output {
		return fmt.Errorf("input and output token cannot be the same: %s", input)
	}

	// We have to compute the ibc voucher string for the osmo coin
	osmoIBCDenom := ibc.ComputeIBCDenom(portID, channelID, "uosmo")
	// We need to get evmDenom from Params to have the code valid also in testnet

	// Check that the input token is evmos or osmo. This constraint will be removed in future
	validInput := []string{stakingDenom, osmoIBCDenom}
	if !slices.Contains(validInput, input) {
		return fmt.Errorf(ErrInputTokenNotSupported, validInput)
	}

	return nil
}

// NewMsgTransfer creates a new MsgTransfer
func NewMsgTransfer(
	sourcePort,
	sourceChannel,
	sender,
	receiver,
	memo string,
	coin sdk.Coin,
	timeoutHeight clienttypes.Height,
) (*transfertypes.MsgTransfer, error) {
	msg := transfertypes.NewMsgTransfer(
		sourcePort,
		sourceChannel,
		coin,
		sender,
		receiver,
		timeoutHeight,
		0,
		memo,
	)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	return msg, nil
}