package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	ethermint "github.com/evmos/ethermint/types"
)

var (
	_ sdk.Msg = &MsgRegisterFeeSplit{}
	_ sdk.Msg = &MsgCancelFeeSplit{}
	_ sdk.Msg = &MsgUpdateFeeSplit{}
)

const (
	TypeMsgRegisterFeeSplit = "register_fee_split"
	TypeMsgCancelFeeSplit   = "cancel_fee_split"
	TypeMsgUpdateFeeSplit   = "update_fee_split"
)

// NewMsgRegisterFeeSplit creates new instance of MsgRegisterFeeSplit
func NewMsgRegisterFeeSplit(
	contract common.Address,
	deployer,
	withdrawer sdk.AccAddress,
	nonces []uint64,
) *MsgRegisterFeeSplit {
	withdrawerAddress := ""
	if withdrawer != nil {
		withdrawerAddress = withdrawer.String()
	}

	return &MsgRegisterFeeSplit{
		ContractAddress:   contract.String(),
		DeployerAddress:   deployer.String(),
		WithdrawerAddress: withdrawerAddress,
		Nonces:            nonces,
	}
}

// Route returns the name of the module
func (msg MsgRegisterFeeSplit) Route() string { return RouterKey }

// Type returns the the action
func (msg MsgRegisterFeeSplit) Type() string { return TypeMsgRegisterFeeSplit }

// ValidateBasic runs stateless checks on the message
func (msg MsgRegisterFeeSplit) ValidateBasic() error {
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
func (msg *MsgRegisterFeeSplit) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgRegisterFeeSplit) GetSigners() []sdk.AccAddress {
	from := sdk.MustAccAddressFromBech32(msg.DeployerAddress)
	return []sdk.AccAddress{from}
}

// NewMsgCancelFeeSplit creates new instance of MsgCancelFeeSplit.
func NewMsgCancelFeeSplit(
	contract common.Address,
	deployer sdk.AccAddress,
) *MsgCancelFeeSplit {
	return &MsgCancelFeeSplit{
		ContractAddress: contract.String(),
		DeployerAddress: deployer.String(),
	}
}

// Route returns the message route for a MsgCancelFeeSplit.
func (msg MsgCancelFeeSplit) Route() string { return RouterKey }

// Type returns the message type for a MsgCancelFeeSplit.
func (msg MsgCancelFeeSplit) Type() string { return TypeMsgCancelFeeSplit }

// ValidateBasic runs stateless checks on the message
func (msg MsgCancelFeeSplit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DeployerAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid deployer address %s", msg.DeployerAddress)
	}

	if err := ethermint.ValidateNonZeroAddress(msg.ContractAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid contract address %s", msg.ContractAddress)
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg *MsgCancelFeeSplit) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgCancelFeeSplit) GetSigners() []sdk.AccAddress {
	funder := sdk.MustAccAddressFromBech32(msg.DeployerAddress)
	return []sdk.AccAddress{funder}
}

// NewMsgUpdateFeeSplit creates new instance of MsgUpdateFeeSplit
func NewMsgUpdateFeeSplit(
	contract common.Address,
	deployer,
	withdraw sdk.AccAddress,
) *MsgUpdateFeeSplit {
	return &MsgUpdateFeeSplit{
		ContractAddress:   contract.String(),
		DeployerAddress:   deployer.String(),
		WithdrawerAddress: withdraw.String(),
	}
}

// Route returns the name of the module
func (msg MsgUpdateFeeSplit) Route() string { return RouterKey }

// Type returns the the action
func (msg MsgUpdateFeeSplit) Type() string { return TypeMsgUpdateFeeSplit }

// ValidateBasic runs stateless checks on the message
func (msg MsgUpdateFeeSplit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DeployerAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid deployer address %s", msg.DeployerAddress)
	}

	if err := ethermint.ValidateNonZeroAddress(msg.ContractAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid contract address %s", msg.ContractAddress)
	}

	if _, err := sdk.AccAddressFromBech32(msg.WithdrawerAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid withdraw address %s", msg.WithdrawerAddress)
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg *MsgUpdateFeeSplit) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgUpdateFeeSplit) GetSigners() []sdk.AccAddress {
	from := sdk.MustAccAddressFromBech32(msg.DeployerAddress)
	return []sdk.AccAddress{from}
}
