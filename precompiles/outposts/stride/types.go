package stride

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
	"math/big"
	"strings"
)

// CreateLiquidStakeEvmosPacket creates a new packet for the liquid staking
func CreateLiquidStakeEvmosPacket(args []interface{}) (common.Address, *big.Int, string, error) {
	if len(args) != 3 {
		return common.Address{}, nil, "", fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 3, len(args))
	}

	erc20Addr, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, nil, "", fmt.Errorf(cmn.ErrInvalidType, "erc20Addr", "", args[0])
	}

	amount, ok := args[1].(*big.Int)
	if !ok {
		return common.Address{}, nil, "", fmt.Errorf(cmn.ErrInvalidType, "amount", "", args[1])
	}

	receiverAddress, ok := args[1].(string)
	if !ok {
		return common.Address{}, nil, "", fmt.Errorf(cmn.ErrInvalidType, "receiverAddress", "", args[2])
	}

	// Check if the receiver address has stride before
	// TODO: This might be unnecessary
	if receiverAddress[:6] != "stride" {
		return common.Address{}, nil, "", fmt.Errorf("receiverAddress is not a stride address")
	}

	// Check if account is a valid bech32 address
	_, err := AccAddressFromBech32(receiverAddress, "stride")
	if err != nil {
		return common.Address{}, nil, "", sdkerrors.ErrInvalidAddress.Wrapf("invalid bech32 address: %s", err)
	}

	return erc20Addr, amount, receiverAddress, nil
}

// NewMsgTransfer creates a new MsgTransfer
func NewMsgTransfer(sourceChannel, senderAddress, receiverAddress, memo string, coin sdk.Coin) (*transfertypes.MsgTransfer, error) {
	// TODO: what are some sensible defaults here
	timeoutHeight := clienttypes.NewHeight(100, 100)

	msg := transfertypes.NewMsgTransfer(
		transfertypes.PortID,
		sourceChannel,
		coin,
		senderAddress,
		receiverAddress,
		timeoutHeight,
		0,
		memo,
	)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	return msg, nil
}

// AccAddressFromBech32 creates an AccAddress from a Bech32 string.
func AccAddressFromBech32(address string, bech32prefix string) (addr sdk.AccAddress, err error) {
	if len(strings.TrimSpace(address)) == 0 {
		return sdk.AccAddress{}, fmt.Errorf("empty address string is not allowed")
	}

	bz, err := sdk.GetFromBech32(address, bech32prefix)
	if err != nil {
		return nil, err
	}

	err = sdk.VerifyAddressFormat(bz)
	if err != nil {
		return nil, err
	}

	return sdk.AccAddress(bz), nil
}
