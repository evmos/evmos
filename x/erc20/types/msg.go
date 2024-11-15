// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	protov2 "google.golang.org/protobuf/proto"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	txsigning "cosmossdk.io/x/tx/signing"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	erc20api "github.com/evmos/evmos/v20/api/evmos/erc20/v1"

	"github.com/ethereum/go-ethereum/common"
)

var (
	_ sdk.Msg              = &MsgMint{}
	_ sdk.Msg              = &MsgBurn{}
	_ sdk.Msg              = &MsgTransferOwnership{}
	_ sdk.Msg              = &MsgConvertERC20{}
	_ sdk.Msg              = &MsgUpdateParams{}
	_ sdk.Msg              = &MsgRegisterERC20{}
	_ sdk.Msg              = &MsgToggleConversion{}
	_ sdk.HasValidateBasic = &MsgConvertERC20{}
	_ sdk.HasValidateBasic = &MsgUpdateParams{}
	_ sdk.HasValidateBasic = &MsgTransferOwnership{}
	_ sdk.HasValidateBasic = &MsgRegisterERC20{}
	_ sdk.HasValidateBasic = &MsgToggleConversion{}
)

const (
	TypeMsgConvertERC20      = "convert_ERC20"
	TypeMsgMint              = "mint"
	TypeMsgBurn              = "burn"
	TypeMsgTransferOwnership = "transfer_ownership"

	AttributeKeyNewOwner = "new_owner"
)

var MsgConvertERC20CustomGetSigner = txsigning.CustomGetSigner{
	MsgType: protov2.MessageName(&erc20api.MsgConvertERC20{}),
	Fn:      erc20api.GetSigners,
}

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

// ValidateBasic does a sanity check of the provided data
func (m *MsgRegisterERC20) ValidateBasic() error {
	for _, addr := range m.Erc20Addresses {
		if !common.IsHexAddress(addr) {
			return errortypes.ErrInvalidAddress.Wrapf("invalid ERC20 contract address: %s", addr)
		}
	}
	return nil
}

// ValidateBasic does a sanity check of the provided data
func (m *MsgToggleConversion) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "Invalid authority address")
	}

	return nil
}

// ValidateBasic does a sanity check of the provided data
func (m *MsgTransferOwnership) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if !common.IsHexAddress(m.Token) {
		return errorsmod.Wrapf(errortypes.ErrInvalidAddress, "invalid ERC20 contract address %s", m.Token)
	}

	if _, err := sdk.AccAddressFromBech32(m.NewOwner); err != nil {
		return errorsmod.Wrap(err, "invalid new owner address")
	}

	return nil
}

// GetSignBytes implements the LegacyMsg interface.
func (m MsgTransferOwnership) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgTransferOwnership message.
func (m MsgTransferOwnership) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Authority)
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
