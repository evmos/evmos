// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/common"
)

var (
	_ sdk.Msg = &MsgConvertCoin{}
	_ sdk.Msg = &MsgConvertERC20{}
	_ sdk.Msg = &MsgUpdateParams{}
)

const (
	TypeMsgConvertCoin  = "convert_coin"
	TypeMsgConvertERC20 = "convert_ERC20"
)

// NewMsgConvertCoin creates a new instance of MsgConvertCoin
func NewMsgConvertCoin(coin sdk.Coin, receiver common.Address, sender sdk.AccAddress) *MsgConvertCoin { //nolint: interfacer
	return &MsgConvertCoin{
		Coin:     coin,
		Receiver: receiver.Hex(),
		Sender:   sender.String(),
	}
}

// Route should return the name of the module
func (msg MsgConvertCoin) Route() string { return RouterKey }

// Type should return the action
func (msg MsgConvertCoin) Type() string { return TypeMsgConvertCoin }

// ValidateBasic runs stateless checks on the message
func (msg MsgConvertCoin) ValidateBasic() error {
	if err := ValidateErc20Denom(msg.Coin.Denom); err != nil {
		if err := ibctransfertypes.ValidateIBCDenom(msg.Coin.Denom); err != nil {
			return err
		}
	}

	if !msg.Coin.Amount.IsPositive() {
		return errorsmod.Wrapf(errortypes.ErrInvalidCoins, "cannot mint a non-positive amount")
	}
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}
	if !common.IsHexAddress(msg.Receiver) {
		return errorsmod.Wrapf(errortypes.ErrInvalidAddress, "invalid receiver hex address %s", msg.Receiver)
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgConvertCoin) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners defines whose signature is required
func (msg MsgConvertCoin) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
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
		return errortypes.Wrap(err, "Invalid authority address")
	}

	if err := m.Params.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes implements the LegacyMsg interface.
func (m MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&m))
}
