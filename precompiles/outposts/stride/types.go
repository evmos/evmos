// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package stride

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/evmos/evmos/v17/utils"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v17/precompiles/common"
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

// AutopilotArgs is the arguments struct for the LiquidStake and RedeemStake methods
type AutopilotArgs struct {
	ChannelID       string         `abi:"channelID"`       // the channel ID for the ICS20 transfer
	Sender          common.Address `abi:"sender"`          // the sender of the liquid stake or redeem transaction
	Receiver        common.Address `abi:"receiver"`        // the receiver of the LSD token or the redeemed token
	Token           common.Address `abi:"token"`           // the token to be liquid staked or redeemed
	Amount          *big.Int       `abi:"amount"`          // the amount to be liquid staked or redeemed
	StrideForwarder string         `abi:"strideForwarder"` // the stride forwarder address
}

// AutopilotPayload is the payload struct for the LiquidStake and RedeemStake method
type AutopilotPayload struct {
	Payload AutopilotArgs
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

// ValidateBasic validates the AutopilotArgs structure and fields
func (a AutopilotArgs) ValidateBasic() error {
	// Check if stride forwarder is a valid bech32 address
	_, err := utils.CreateAccAddressFromBech32(a.StrideForwarder, StrideBech32Prefix)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid stride bech32 address: %s", err)
	}

	if a.Amount.Sign() <= 0 {
		return fmt.Errorf(ErrZeroOrNegativeAmount)
	}

	return nil
}

// parseAutopilotArgs parses the arguments from the Liquid Stake and for Redeem Stake method calls
func parseAutopilotArgs(method *abi.Method, args []interface{}) (AutopilotArgs, error) {
	if len(args) != 1 {
		return AutopilotArgs{}, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	var autopilotPayload AutopilotPayload
	if err := method.Inputs.Copy(&autopilotPayload, args); err != nil {
		return AutopilotArgs{}, fmt.Errorf("error while unpacking args to AutopilotArgs struct: %s", err)
	}

	// Validate the AutopilotArgs struct
	if err := autopilotPayload.Payload.ValidateBasic(); err != nil {
		return AutopilotArgs{}, err
	}

	return autopilotPayload.Payload, nil
}

// CreateMemo creates the memo for the StakeIBC actions - LiquidStake and RedeemStake.
func CreateMemo(action, strideForwarder, receiver string) (string, error) {
	// Create a new instance of the struct and populate it
	data := &RawPacketMetadata{
		Autopilot: &Autopilot{
			Receiver: strideForwarder,
			StakeIBC: &StakeIBCPacketMetadata{
				Action:      action,
				IBCReceiver: receiver,
			},
		},
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
