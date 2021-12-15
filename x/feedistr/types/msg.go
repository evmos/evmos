package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	ethermint "github.com/tharsis/ethermint/types"
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
	if err := ethermint.ValidateAddress(msg.ContractAddress); err != nil {
		return sdkerrors.Wrap(err, "smart contract address")
	}
	if err := ethermint.ValidateAddress(msg.DeployerAddress); err != nil {
		return sdkerrors.Wrap(err, "deployer address")
	}

	contractAddr := common.HexToAddress(msg.ContractAddress)
	deployerAddr := common.HexToAddress(msg.DeployerAddress)

	expContractAddr := crypto.CreateAddress(deployerAddr, msg.Nonce)
	if expContractAddr != contractAddr {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"expected contract address %s, got %s",
			expContractAddr, msg.ContractAddress,
		)
	}

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

// NewMsgRegisterContract creates a new instance of MsgRegisterContract
func NewMsgUpdateWithdawAddress(contractAddr, withdrawAddress, newWithdrawAddress common.Address) *MsgUpdateWithdawAddress { // nolint: interfacer
	return &MsgUpdateWithdawAddress{
		ContractAddress:      contractAddr.Hex(),
		NewWithdrawalAddress: newWithdrawAddress.Hex(),
		WithdrawalAddress:    withdrawAddress.Hex(),
	}
}

// Route should return the name of the module
func (msg MsgUpdateWithdawAddress) Route() string { return RouterKey }

// Type should return the action
func (msg MsgUpdateWithdawAddress) Type() string { return TypeMsgUpdateWithdawAddress }

// ValidateBasic runs stateless checks on the message
func (msg MsgUpdateWithdawAddress) ValidateBasic() error {
	if err := ethermint.ValidateAddress(msg.ContractAddress); err != nil {
		return sdkerrors.Wrap(err, "smart contract address")
	}
	if err := ethermint.ValidateAddress(msg.NewWithdrawalAddress); err != nil {
		return sdkerrors.Wrap(err, "new withdrawal address")
	}
	if err := ethermint.ValidateAddress(msg.WithdrawalAddress); err != nil {
		return sdkerrors.Wrap(err, "withdrawal address")
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg *MsgUpdateWithdawAddress) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgUpdateWithdawAddress) GetSigners() []sdk.AccAddress {
	addr := sdk.AccAddress(common.HexToAddress(msg.WithdrawalAddress).Bytes())
	return []sdk.AccAddress{addr}
}
