package types

import (
	"math/big"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgCallEVM{}
)

const (
	TypeMsgCallEVM = "call_evm"
)

// NewMsgCallEVM creates a new instance of MsgCallEVM
func NewMsgCallEVM(amount string, denom string, packet *IBCEVMPacketData) *MsgCallEVM {
	return &MsgCallEVM{
		Amount: amount,
		Denom:  denom,
		Packet: packet,
	}
}

// Route should return the name of the module
func (msg MsgCallEVM) Route() string { return RouterKey }

// Type should return the action
func (msg MsgCallEVM) Type() string { return TypeMsgCallEVM }

// ValidateBasic runs stateless checks on the message
func (msg MsgCallEVM) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Packet.Sender)
	if err != nil {
		return sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "Invalid AccAddress")
	}

	amount := new(big.Int)
	amount.SetString(msg.Amount, 10)

	// TODO: Should not be negative but can it be zero ?
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return sdkerrors.Wrapf(errortypes.ErrInvalidRequest, "Invalid amount")
	}

	return nil
}

// GetSignBytes encodes the message for signing
// TODO: Check if we really need the amino codec here (currently mirrored from )
func (msg MsgCallEVM) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for the Tx
func (msg *MsgCallEVM) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(msg.Packet.Sender)
	return []sdk.AccAddress{addr}
}
