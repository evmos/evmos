package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
)

var (
	_ sdk.Msg = &MsgRegisterContract{}
	_ sdk.Msg = &MsgUpdateWithdawAddress{}
)

const (
	TypeMsgRegisterContract     = "register_contract"
	TypeMsgUpdateWithdawAddress = "update_withdraw_address"
)

// NewMsgRegisterContract creates a new instance of MsgRegisterContract
func NewMsgRegisterContract(contractAddr, deployerAddr common.Address, nonce uint64) *MsgRegisterContract { // nolint: interfacer
	return &MsgRegisterContract{
		ContractAddress: contractAddr.Hex(),
		Nonce:           nonce,
		DeployerAddress: deployerAddr.Hex(),
	}
}

// Route should return the name of the module
func (msg MsgRegisterContract) Route() string { return RouterKey }

// Type should return the action
func (msg MsgRegisterContract) Type() string { return TypeMsgRegisterContract }

// ValidateBasic runs stateless checks on the message
func (msg MsgRegisterContract) ValidateBasic() error {
	// TODO
	return nil
}

// GetSignBytes encodes the message for signing
func (msg *MsgRegisterContract) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgRegisterContract) GetSigners() []sdk.AccAddress {
	addr := sdk.AccAddress(common.HexToAddress(msg.DeployerAddress).Bytes())
	return []sdk.AccAddress{addr}
}
