package stride

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	cmn "github.com/evmos/evmos/v14/precompiles/common"
	"strings"
)

// CreateLiquidStakeEvmosPacket creates a new packet for the liquid staking
func CreateLiquidStakeEvmosPacket(args []interface{}, bondDenom string) (sdk.Coin, string, error) {
	if len(args) != 2 {
		return sdk.Coin{}, "", fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	coin, ok := args[0].(cmn.Coin)
	if !ok {
		return sdk.Coin{}, "", fmt.Errorf(cmn.ErrInvalidType, "amount", cmn.Coin{}, args[0])
	}

	if coin.Denom != bondDenom {
		return sdk.Coin{}, "", fmt.Errorf(cmn.ErrInvalidDenom, "aevmos")
	}

	// Convert our common Coin into an SDK Coin
	sdkCoin := sdk.NewCoin(coin.Denom, sdk.NewIntFromBigInt(coin.Amount))

	receiverAddress, ok := args[1].(string)
	if !ok {
		return sdk.Coin{}, "", fmt.Errorf(cmn.ErrInvalidType, "receiverAddress", "", args[1])
	}

	// Check if the receiver address has stride before
	if receiverAddress[:6] != "stride" {
		return sdk.Coin{}, "", fmt.Errorf("receiverAddress is not a stride address")
	}

	// Check if account is a valid bech32 address
	_, err := AccAddressFromBech32(receiverAddress, "stride")
	if err != nil {
		return sdk.Coin{}, "", sdkerrors.ErrInvalidAddress.Wrapf("invalid bech32 address: %s", err)
	}

	return sdkCoin, receiverAddress, nil
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
