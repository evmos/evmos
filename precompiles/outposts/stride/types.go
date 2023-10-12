package stride

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
)

// StakeIBCPacketMetadata metadata info specific to StakeIBC (e.g. 1-click liquid staking)
type StakeIBCPacketMetadata struct {
	Action        string `json:"action"`
	StrideAddress string
}

// RawPacketMetadata is the raw packet metadata used to construct a JSON string
type RawPacketMetadata struct {
	Autopilot *struct {
		Receiver string                  `json:"receiver"`
		StakeIBC *StakeIBCPacketMetadata `json:"stakeibc,omitempty"`
	} `json:"autopilot"`
}

// ParseLiquidStakeArgs parses the arguments from the Liquid Stake method call
func ParseLiquidStakeArgs(args []interface{}) (common.Address, common.Address, *big.Int, string, error) {
	if len(args) != 4 {
		return common.Address{}, common.Address{}, nil, "", fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 4, len(args))
	}

	sender, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, nil, "", fmt.Errorf(cmn.ErrInvalidType, "sender", "", args[0])
	}

	token, ok := args[1].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, nil, "", fmt.Errorf(cmn.ErrInvalidType, "token", "", args[1])
	}

	amount, ok := args[2].(*big.Int)
	if !ok {
		return common.Address{}, common.Address{}, nil, "", fmt.Errorf(cmn.ErrInvalidType, "amount", "", args[2])
	}

	receiver, ok := args[3].(string)
	if !ok {
		return common.Address{}, common.Address{}, nil, "", fmt.Errorf(cmn.ErrInvalidType, "receiver", "", args[3])
	}

	// Check if the receiver address has stride before
	// TODO: This might be unnecessary
	if receiver[:6] != "stride" {
		return common.Address{}, common.Address{}, nil, "", fmt.Errorf("receiver is not a stride address")
	}

	// Check if account is a valid bech32 address

	_, err := AccAddressFromBech32(receiver, "stride")
	if err != nil {
		return common.Address{}, common.Address{}, nil, "", sdkerrors.ErrInvalidAddress.Wrapf("invalid stride bech32 address: %s", err)
	}

	return sender, token, amount, receiver, nil
}

// NewMsgTransfer creates a new MsgTransfer
func NewMsgTransfer(sourceChannel, senderAddress, receiverAddress, memo string, coin sdk.Coin) (*transfertypes.MsgTransfer, error) {
	// TODO: what are some sensible defaults here
	timeoutHeight := clienttypes.NewHeight(100, 100)

	msg := transfertypes.NewMsgTransfer(
		transfertypes.PortID,
		sourceChannel,
		coin,
		senderAddress,
		receiverAddress,
		timeoutHeight,
		0,
		memo,
	)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	return msg, nil
}

// AccAddressFromBech32 creates an AccAddress from a Bech32 string.
func AccAddressFromBech32(address string, bech32prefix string) (addr sdk.AccAddress, err error) {
	if len(strings.TrimSpace(address)) == 0 {
		return sdk.AccAddress{}, fmt.Errorf("empty address string is not allowed")
	}

	bz, err := sdk.GetFromBech32(address, bech32prefix)
	if err != nil {
		return nil, err
	}

	err = sdk.VerifyAddressFormat(bz)
	if err != nil {
		return nil, err
	}

	return sdk.AccAddress(bz), nil
}

// createLiquidStakeMemo creates the memo for the LiquidStake packet
func (p Precompile) createMemo(action, receiverAddress string) string {
	// Create a new instance of the struct and populate it
	data := &RawPacketMetadata{
		Autopilot: &struct {
			Receiver string                  `json:"receiver"`
			StakeIBC *StakeIBCPacketMetadata `json:"stakeibc,omitempty"`
		}{
			Receiver: receiverAddress,
			StakeIBC: &StakeIBCPacketMetadata{
				Action: action,
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
	return string(jsonBytes)
}
