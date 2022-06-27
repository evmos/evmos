package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	ethermint "github.com/evmos/ethermint/types"
)

var (
	_ sdk.Msg = &MsgRegisterFee{}
	_ sdk.Msg = &MsgCancelFee{}
	_ sdk.Msg = &MsgUpdateFee{}
)

const (
	TypeMsgRegisterFee = "register_fee"
	TypeMsgCancelFee   = "cancel_fee"
	TypeMsgUpdateFee   = "update_fee"
)

// NewMsgRegisterFee creates new instance of MsgRegisterFee
func NewMsgRegisterFee(
	contract common.Address,
	deployer,
	withdrawal sdk.AccAddress,
	nonces []uint64,
) *MsgRegisterFee {
	var withdrawAddr string
	if withdrawal != nil {
		withdrawAddr = withdrawal.String()
	}

	return &MsgRegisterFee{
		ContractAddress:   contract.String(),
		DeployerAddress:   deployer.String(),
		WithdrawerAddress: withdrawAddr,
		Nonces:            nonces,
	}
}

// Route returns the name of the module
func (msg MsgRegisterFee) Route() string { return RouterKey }

// Type returns the the action
func (msg MsgRegisterFee) Type() string { return TypeMsgRegisterFee }

// ValidateBasic runs stateless checks on the message
func (msg MsgRegisterFee) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DeployerAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid deployer address %s", msg.DeployerAddress)
	}

	if err := ethermint.ValidateNonZeroAddress(msg.ContractAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid contract address %s", msg.ContractAddress)
	}

	if msg.WithdrawerAddress != "" {
		if _, err := sdk.AccAddressFromBech32(msg.WithdrawerAddress); err != nil {
			return sdkerrors.Wrapf(err, "invalid withdraw address %s", msg.WithdrawerAddress)
		}
		if msg.DeployerAddress == msg.WithdrawerAddress {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "please specify an empty withdraw address instead of setting the deployer address")
		}
	}

	if len(msg.Nonces) < 1 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid nonces - empty array")
	}

	if len(msg.Nonces) > 20 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid nonces - array length must be less than 20")
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg *MsgRegisterFee) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgRegisterFee) GetSigners() []sdk.AccAddress {
	from := sdk.MustAccAddressFromBech32(msg.DeployerAddress)
	return []sdk.AccAddress{from}
}

// NewMsgCancelFee creates new instance of MsgCancelFee.
func NewMsgCancelFee(
	contract common.Address,
	deployer sdk.AccAddress,
) *MsgCancelFee {
	return &MsgCancelFee{
		ContractAddress: contract.String(),
		DeployerAddress: deployer.String(),
	}
}

// Route returns the message route for a MsgCancelFee.
func (msg MsgCancelFee) Route() string { return RouterKey }

// Type returns the message type for a MsgCancelFee.
func (msg MsgCancelFee) Type() string { return TypeMsgCancelFee }

// ValidateBasic runs stateless checks on the message
func (msg MsgCancelFee) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DeployerAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid deployer address %s", msg.DeployerAddress)
	}

	if err := ethermint.ValidateNonZeroAddress(msg.ContractAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid contract address %s", msg.ContractAddress)
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg *MsgCancelFee) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgCancelFee) GetSigners() []sdk.AccAddress {
	funder := sdk.MustAccAddressFromBech32(msg.DeployerAddress)
	return []sdk.AccAddress{funder}
}

// NewMsgUpdateFee creates new instance of MsgUpdateFee
func NewMsgUpdateFee(
	contract common.Address,
	deployer,
	withdraw sdk.AccAddress,
) *MsgUpdateFee {
	return &MsgUpdateFee{
		ContractAddress:   contract.String(),
		DeployerAddress:   deployer.String(),
		WithdrawerAddress: withdraw.String(),
	}
}

// Route returns the name of the module
func (msg MsgUpdateFee) Route() string { return RouterKey }

// Type returns the the action
func (msg MsgUpdateFee) Type() string { return TypeMsgUpdateFee }

// ValidateBasic runs stateless checks on the message
func (msg MsgUpdateFee) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DeployerAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid deployer address %s", msg.DeployerAddress)
	}

	if err := ethermint.ValidateNonZeroAddress(msg.ContractAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid contract address %s", msg.ContractAddress)
	}

	if _, err := sdk.AccAddressFromBech32(msg.WithdrawerAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid withdraw address %s", msg.WithdrawerAddress)
	}

	if msg.DeployerAddress == msg.WithdrawerAddress {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "withdraw address must be different that deployer address: withdraw %s, deployer %s", msg.WithdrawerAddress, msg.DeployerAddress)
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg *MsgUpdateFee) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgUpdateFee) GetSigners() []sdk.AccAddress {
	from := sdk.MustAccAddressFromBech32(msg.DeployerAddress)
	return []sdk.AccAddress{from}
}
