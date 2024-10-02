// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgBid{}
	_ sdk.Msg = &MsgDepositCoin{}
	_ sdk.Msg = &MsgUpdateParams{}
)

const (
	TypeMsgBid          = "bid"
	TypeMsgDeposit      = "deposit"
	TypeMsgUpdateParams = "update_params"
)

func (msg MsgBid) Route() string {
	return RouterKey
}

func (msg MsgBid) Type() string {
	return TypeMsgBid
}

// ValidateBasic runs stateless checks on the message
func (msg MsgBid) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errors.Wrap(err, "invalid sender address")
	}

	if msg.Amount.Denom != BidDenom {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "denom should be %s", BidDenom)
	}
	if msg.Amount.IsZero() || msg.Amount.IsNegative() {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "amount cannot be zero or negative")
	}
	return nil
}

// GetSigners defines whose signature is required
func (msg MsgBid) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgBid) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

func (msg MsgDepositCoin) Route() string {
	return RouterKey
}

func (msg MsgDepositCoin) Type() string {
	return TypeMsgDeposit
}

// ValidateBasic runs stateless checks on the message
func (msg MsgDepositCoin) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errors.Wrap(err, "invalid sender address")
	}

	if msg.Amount.Denom == BidDenom {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "denom cannot be be %s", BidDenom)
	}
	if msg.Amount.IsZero() || msg.Amount.IsNegative() {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "amount cannot be zero or negative")
	}
	return nil
}

// GetSigners defines whose signature is required
func (msg MsgDepositCoin) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgDepositCoin) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

func (msg MsgUpdateParams) Route() string {
	return RouterKey
}

func (msg MsgUpdateParams) Type() string {
	return TypeMsgUpdateParams
}

// ValidateBasic runs stateless checks on the message
func (msg MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errors.Wrap(err, "invalid authority address")
	}

	return msg.Params.Validate()
}

// GetSigners defines whose signature is required
func (msg MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}
