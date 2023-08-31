package stride

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/evmos/evmos/v14/precompiles/common"
)

// CreateLiquidStakeEvmosPacket creates a new packet for the liquid staking
func CreateLiquidStakeEvmosPacket(args []interface{}, bondDenom string) (sdk.Coin, string, error) {
	if len(args) != 2 {
		return sdk.Coin{}, "", fmt.Errorf("too many arguments")
	}

	coin, ok := args[0].(common.Coin)
	if !ok {
		return sdk.Coin{}, "", fmt.Errorf("amount is not a big.Int")
	}

	if coin.Denom != bondDenom {
		return sdk.Coin{}, "", fmt.Errorf("denom is not aevmos")
	}

	// Convert our common Coin into an SDK Coin
	sdkCoin := sdk.NewCoin(coin.Denom, sdk.NewIntFromBigInt(coin.Amount))

	receiverAddress, ok := args[1].(string)
	if !ok {
		return sdk.Coin{}, "", fmt.Errorf("receiverAddress is not a string")
	}

	// TODO: check with the prefix of the receiver chain (Stride in this case)
	// Check if the receiver address has stride before
	if receiverAddress[:5] != "stride" {
		return sdk.Coin{}, "", fmt.Errorf("receiverAddress is not a stride address")
	}

	// TODO: check if the amount is zero and error out if it is

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
