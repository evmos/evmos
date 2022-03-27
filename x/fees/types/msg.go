package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

var (
	_ sdk.Msg = &MsgRegisterFeeContract{}
	_ sdk.Msg = &MsgCancelFeeContract{}
	_ sdk.Msg = &MsgUpdateFeeContract{}
)

const (
	TypeMsgRegisterFeeContract = "register_fee_contract"
	TypeMsgCancelFeeContract   = "cancel_fee_contract"
	TypeMsgUpdateFeeContract   = "update_fee_contract"
)

// NewMsgRegisterFeeContract creates new instance of MsgRegisterFeeContract
func NewMsgRegisterFeeContract(
	fromAddr sdk.AccAddress,
	contract string,
	deploymentHash string,
	withdrawAddress sdk.AccAddress,
) *MsgRegisterFeeContract {
	return &MsgRegisterFeeContract{
		FromAddress:     fromAddr.String(),
		Contract:        contract,
		DeploymentHash:  deploymentHash,
		WithdrawAddress: withdrawAddress.String(),
	}
}

// Route returns the name of the module
func (msg MsgRegisterFeeContract) Route() string { return RouterKey }

// Type returns the the action
func (msg MsgRegisterFeeContract) Type() string { return TypeMsgRegisterFeeContract }

// ValidateBasic runs stateless checks on the message
func (msg MsgRegisterFeeContract) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid from address")
	}

	if !common.IsHexAddress(msg.Contract) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid contract address %s", msg.Contract)
	}

	if _, err := sdk.AccAddressFromBech32(msg.WithdrawAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid withdrawal address address")
	}

	// TODO check deployment hash
	return nil
}

// GetSignBytes encodes the message for signing
func (msg *MsgRegisterFeeContract) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgRegisterFeeContract) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{from}
}

// NewMsgClawbackcreates new instance of MsgClawback. The dest_address may be
// nil - defaulting to the funder.
func NewMsgCancelFeeContract(fromAddr sdk.AccAddress, contract string) *MsgCancelFeeContract {
	return &MsgCancelFeeContract{
		Contract:    contract,
		FromAddress: fromAddr.String(),
	}
}

// Route returns the message route for a MsgCancelFeeContract.
func (msg MsgCancelFeeContract) Route() string { return RouterKey }

// Type returns the message type for a MsgCancelFeeContract.
func (msg MsgCancelFeeContract) Type() string { return TypeMsgCancelFeeContract }

// ValidateBasic runs stateless checks on the message
func (msg MsgCancelFeeContract) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.GetFromAddress()); err != nil {
		return sdkerrors.Wrapf(err, "invalid from address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.GetContract()); err != nil {
		return sdkerrors.Wrapf(err, "invalid contract address")
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg *MsgCancelFeeContract) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgCancelFeeContract) GetSigners() []sdk.AccAddress {
	funder, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{funder}
}

// NewMsgUpdateFeeContract creates new instance of MsgUpdateFeeContract
func NewMsgUpdateFeeContract(
	fromAddr sdk.AccAddress,
	contract string,
	withdrawAddress sdk.AccAddress,
) *MsgUpdateFeeContract {
	return &MsgUpdateFeeContract{
		FromAddress:     fromAddr.String(),
		Contract:        contract,
		WithdrawAddress: withdrawAddress.String(),
	}
}

// Route returns the name of the module
func (msg MsgUpdateFeeContract) Route() string { return RouterKey }

// Type returns the the action
func (msg MsgUpdateFeeContract) Type() string { return TypeMsgUpdateFeeContract }

// ValidateBasic runs stateless checks on the message
func (msg MsgUpdateFeeContract) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.FromAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid from address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Contract); err != nil {
		return sdkerrors.Wrapf(err, "invalid contract address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.WithdrawAddress); err != nil {
		return sdkerrors.Wrapf(err, "invalid withdrawal address address")
	}

	// TODO check deployment hash
	return nil
}

// GetSignBytes encodes the message for signing
func (msg *MsgUpdateFeeContract) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgUpdateFeeContract) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{from}
}
