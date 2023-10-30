// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package stride

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/evmos/evmos/v15/utils"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
)

const (
	// StrideBech32Prefix is the Bech32 prefix for Stride addresses
	StrideBech32Prefix = "stride"
)

// EventLiquidStake is the event type emitted on a liquidStake transaction
type EventLiquidStake struct {
	Sender common.Address
	Token  common.Address
	Amount *big.Int
}

// EventRedeem is the event type emitted on a redeem transaction
type EventRedeem struct {
	Sender          common.Address
	Token           common.Address
	Receiver        common.Address
	StrideForwarder string
	Amount          *big.Int
}

// StakeIBCPacketMetadata metadata info specific to StakeIBC (e.g. 1-click liquid staking).
// Used to create the memo field for the ICS20 transfer corresponding to Autopilot LiquidStake.
type StakeIBCPacketMetadata struct {
	Action      string `json:"action"`
	IBCReceiver string `json:"ibcreceiver,omitempty"`
}

// Autopilot defines the receiver and IBC packet metadata info specific to the
// Stride Autopilot liquid staking behavior
type Autopilot struct {
	Receiver string                  `json:"receiver"`
	StakeIBC *StakeIBCPacketMetadata `json:"stakeibc,omitempty"`
}

// RawPacketMetadata is the raw packet metadata used to construct a JSON string
type RawPacketMetadata struct {
	Autopilot *Autopilot `json:"autopilot"`
}

// ValidateBasic validates the RawPacketMetadata structure and fields
func (r RawPacketMetadata) ValidateBasic() error {
	if r.Autopilot.StakeIBC.Action == "" {
		return fmt.Errorf(ErrEmptyAutopilotAction)
	}

	if r.Autopilot.Receiver == "" {
		return fmt.Errorf(ErrEmptyReceiver)
	}

	if r.Autopilot.StakeIBC.Action == RedeemStakeAction && r.Autopilot.StakeIBC.IBCReceiver == "" {
		return fmt.Errorf(ErrRedeemStakeEmptyIBCReceiver)
	}

	return nil
}

// parseLiquidStakeArgs parses the arguments from the Liquid Stake method call
func parseLiquidStakeArgs(args []interface{}) (common.Address, common.Address, *big.Int, string, error) {
	if len(args) != 4 {
		return common.Address{}, common.Address{}, nil, "", fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 4, len(args))
	}

	sender, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, nil, "", fmt.Errorf(cmn.ErrInvalidType, "sender", common.Address{}, args[0])
	}

	token, ok := args[1].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, nil, "", fmt.Errorf(cmn.ErrInvalidType, "token", common.Address{}, args[1])
	}

	amount, ok := args[2].(*big.Int)
	if !ok {
		return common.Address{}, common.Address{}, nil, "", fmt.Errorf(cmn.ErrInvalidType, "amount", &big.Int{}, args[2])
	}

	receiver, ok := args[3].(string)
	if !ok {
		return common.Address{}, common.Address{}, nil, "", fmt.Errorf(cmn.ErrInvalidType, "receiver", "", fmt.Sprintf("%T", args[3]))
	}

	// Check if the receiver address has stride before
	if receiver[:6] != StrideBech32Prefix {
		return common.Address{}, common.Address{}, nil, "", fmt.Errorf("receiver is not a stride address")
	}

	// Check if account is a valid bech32 address
	_, err := utils.CreateAccAddressFromBech32(receiver, StrideBech32Prefix)
	if err != nil {
		return common.Address{}, common.Address{}, nil, "", sdkerrors.ErrInvalidAddress.Wrapf("invalid stride bech32 address: %s", err)
	}

	return sender, token, amount, receiver, nil
}

// parseRedeemStakeArgs parses the arguments from the Redeem Stake method call
func parseRedeemStakeArgs(args []interface{}) (common.Address, common.Address, common.Address, string, *big.Int, error) {
	if len(args) != 5 {
		return common.Address{}, common.Address{}, common.Address{}, "", nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 5, len(args))
	}

	sender, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, "", nil, fmt.Errorf(cmn.ErrInvalidType, "sender", common.Address{}, args[0])
	}

	receiver, ok := args[1].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, "", nil, fmt.Errorf(cmn.ErrInvalidType, "receiver", common.Address{}, args[1])
	}

	token, ok := args[2].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, "", nil, fmt.Errorf(cmn.ErrInvalidType, "token", common.Address{}, args[2])
	}

	amount, ok := args[3].(*big.Int)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, "", nil, fmt.Errorf(cmn.ErrInvalidType, "amount", &big.Int{}, args[3])
	}

	strideForwarder, ok := args[4].(string)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, "", nil, fmt.Errorf(cmn.ErrInvalidType, "strideForwardeer", "", fmt.Sprintf("%T", args[4]))
	}

	// Check if the receiver address has stride before
	if strideForwarder[:6] != StrideBech32Prefix {
		return common.Address{}, common.Address{}, common.Address{}, "", nil, fmt.Errorf("receiver is not a stride address")
	}

	// Check if account is a valid bech32 address
	_, err := utils.CreateAccAddressFromBech32(strideForwarder, StrideBech32Prefix)
	if err != nil {
		return common.Address{}, common.Address{}, common.Address{}, "", nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid stride bech32 address: %s", err)
	}

	return sender, receiver, token, strideForwarder, amount, nil
}

// CreateMemo creates the memo for the StakeIBC actions - LiquidStake and RedeemStake.
func CreateMemo(action, strideForwarder, receiver string) (string, error) {
	// Create a new instance of the struct and populate it
	data := &RawPacketMetadata{
		Autopilot: &Autopilot{
			Receiver: strideForwarder,
			StakeIBC: &StakeIBCPacketMetadata{
				Action: action,
			},
		},
	}

	// Populate the IBC Receiver field if the action is RedeemStake
	if action == RedeemStakeAction {
		data.Autopilot.StakeIBC.IBCReceiver = receiver
	}

	if err := data.ValidateBasic(); err != nil {
		return "", err
	}

	// Convert the struct to a JSON string
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", sdkerrors.ErrJSONMarshal.Wrap("autopilot packet")
	}

	return string(jsonBytes), nil
}
