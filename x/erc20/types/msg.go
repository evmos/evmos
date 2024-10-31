// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ethereum/go-ethereum/common"
)

var (
	_ sdk.Msg = &MsgConvertERC20{}
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgMint{}
	_ sdk.Msg = &MsgBurn{}
)

const (
	TypeMsgConvertERC20 = "convert_ERC20"
	TypeMsgMint         = "mint"
	TypeMsgBurn         = "burn"
	TypeMsgTransferOwnership = "transfer_ownership"
)

// NewMsgConvertERC20 creates a new instance of MsgConvertERC20
func NewMsgConvertERC20(amount math.Int, receiver sdk.AccAddress, contract, sender common.Address) *MsgConvertERC20 { //nolint: interfacer
	return &MsgConvertERC20{
		ContractAddress: contract.String(),
		Amount:          amount,
		Receiver:        receiver.String(),
		Sender:          sender.Hex(),
	}
}

// Route should return the name of the module
func (msg MsgConvertERC20) Route() string { return RouterKey }

// Type should return the action
func (msg MsgConvertERC20) Type() string { return TypeMsgConvertERC20 }

// ValidateBasic runs stateless checks on the message
func (msg MsgConvertERC20) ValidateBasic() error {
	if !common.IsHexAddress(msg.ContractAddress) {
		return errorsmod.Wrapf(errortypes.ErrInvalidAddress, "invalid contract hex address '%s'", msg.ContractAddress)
	}
	if !msg.Amount.IsPositive() {
		return errorsmod.Wrapf(errortypes.ErrInvalidCoins, "cannot mint a non-positive amount")
	}
	_, err := sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return errorsmod.Wrap(err, "invalid receiver address")
	}
	if !common.IsHexAddress(msg.Sender) {
		return errorsmod.Wrapf(errortypes.ErrInvalidAddress, "invalid sender hex address %s", msg.Sender)
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgConvertERC20) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners defines whose signature is required
func (msg MsgConvertERC20) GetSigners() []sdk.AccAddress {
	addr := common.HexToAddress(msg.Sender)
	return []sdk.AccAddress{addr.Bytes()}
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check of the provided data
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "Invalid authority address")
	}

	return m.Params.Validate()
}

// GetSignBytes implements the LegacyMsg interface.
func (m MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgMint message.
func (m MsgMint) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check of the provided data
func (m MsgMint) ValidateBasic() error {
	if !common.IsHexAddress(m.ContractAddress) {
		return errorsmod.Wrapf(errortypes.ErrInvalidAddress, "invalid contract hex address '%s'", m.ContractAddress)
	}

	if !m.Amount.IsPositive() {
		return errorsmod.Wrapf(errortypes.ErrInvalidCoins, "cannot mint a non-positive amount")
	}

	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(m.To); err != nil {
		return errorsmod.Wrap(err, "invalid receiver address")
	}

	return nil
}

// Route returns the message route for a MsgMint
func (m MsgMint) Route() string { return RouterKey }

// Type returns the message type for a MsgMint
func (m MsgMint) Type() string { return TypeMsgMint }

// GetSigners returns the expected signers for a MsgBurn message.
func (m MsgBurn) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check of the provided data	
func (m MsgBurn) ValidateBasic() error {
	if !common.IsHexAddress(m.ContractAddress) {
		return errorsmod.Wrapf(errortypes.ErrInvalidAddress, "invalid contract hex address '%s'", m.ContractAddress)
	}

	if !m.Amount.IsPositive() {
		return errorsmod.Wrapf(errortypes.ErrInvalidCoins, "cannot burn a non-positive amount")
	}

	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	return nil
}

// Route returns the message route for a MsgBurn
func (m MsgBurn) Route() string { return RouterKey }

// Type returns the message type for a MsgBurn
func (m MsgBurn) Type() string { return TypeMsgBurn }

