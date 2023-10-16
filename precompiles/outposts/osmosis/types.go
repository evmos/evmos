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

type Twap struct {
	SlippagePercentage string `json:"slippage_percentage"`
	WindowSeconds      int64 `json:"window_seconds"`
}

type Slippage struct {
	Twap Twap `json:"twap"`
}

type OsmosisSwap struct {
	OutputDenom      string   `json:"output_denom"`
	Slippage         Slippage `json:"slippage"`
	Receiver         string   `json:"receiver"`
	OnFailedDelivery string   `json:"on_failed_delivery"`
	NextMemo         string   `json:"next_memo"`
}

type Msg struct {
	OsmosisSwap OsmosisSwap `json:"osmosis_swap,omitempty"`
}

type Memo struct {
	Contract string `json:"contract"`
	Msg      Msg    `json:"msg"`
}

type RawPacketMetadata struct {
	Memo Memo `json:"memo"`
}

func CreateMemo(outputDenom, receiver, contract, slippage_percentage string, window_seconds int64) (string, error) {
	data := &RawPacketMetadata{
		Memo{
			Contract: contract,
			Msg: Msg{
				OsmosisSwap: OsmosisSwap{
					OutputDenom: outputDenom,
					Slippage: Slippage{
						Twap{
							SlippagePercentage: slippage_percentage, 
							WindowSeconds:      window_seconds, 
						},
					},
					Receiver:         receiver,
					OnFailedDelivery: "do_nothing",
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

// ParseSwapPacketData parses the packet data for the Osmosis swap function.
func ParseSwapPacketData(args []interface{}) (sender common.Address, input common.Address, output common.Address, amount *big.Int, receiver string, err error) {
	if len(args) != 5 {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 5, len(args))
	}

	sender, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", fmt.Errorf("invalid sender address: %v", args[0])
	}

	input, ok = args[1].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", fmt.Errorf("invalid input denom: %v", args[1])
	}

	output, ok = args[2].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", fmt.Errorf("invalid output denom: %v", args[2])
	}

	amount, ok = args[3].(*big.Int)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", fmt.Errorf("invalid amount: %v", args[3])
	}

	receiver, ok = args[4].(string)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, nil, "", fmt.Errorf("invalid receiver address: %v", args[4])
	}

	return sender, input, output, amount, receiver, nil
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
func NewMsgTransfer(sourcePort, sourceChannel, sender, receiver, memo string, coin sdk.Coin) (*transfertypes.MsgTransfer, error) {
	// TODO: what are some sensible defaults here
	timeoutHeight := clienttypes.NewHeight(100, 100)
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
