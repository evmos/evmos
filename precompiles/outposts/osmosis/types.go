package osmosis

import (
	"cosmossdk.io/math"
	"fmt"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/ethereum/go-ethereum/common"
	erc20types "github.com/evmos/evmos/v14/x/erc20/types"
)

// CreateSwapPacketData creates the packet data for the Osmosis swap function.
func CreateSwapPacketData(args []interface{}, ctx sdk.Context, bankKeeper erc20types.BankKeeper) (*big.Int, string, string, string, error) {
	if len(args) != 4 {
		return nil, "", "", "", fmt.Errorf("invalid number of arguments: %d", len(args))
	}

	amount, ok := args[0].(*big.Int)
	if !ok {
		return nil, "", "", "", fmt.Errorf("invalid amount: %v", args[0])
	}

	inputContract, ok := args[1].(common.Address)
	if !ok {
		return nil, "", "", "", fmt.Errorf("invalid input denom: %v", args[1])
	}

	metadata, found := bankKeeper.GetDenomMetaData(ctx, inputContract.String())
	if !found {
		return nil, "", "", "", fmt.Errorf("invalid input denom: %v", inputContract.String())
	}

	outputDenom, ok := args[2].(string)
	if !ok {
		return nil, "", "", "", fmt.Errorf("invalid output denom: %v", args[2])
	}

	receiverAddress, ok := args[3].(string)
	if !ok {
		return nil, "", "", "", fmt.Errorf("invalid receiver address: %v", args[3])
	}

	prefix, _, err := bech32.DecodeAndConvert(receiverAddress)
	if err != nil {
		return nil, "", "", "", fmt.Errorf("invalid receiver address: %v", err)
	}

	fmt.Println(prefix)

	return amount, metadata.Base, outputDenom, receiverAddress, nil
}

// NewMsgTransfer returns a new transfer message from the given arguments.
func NewMsgTransfer(denom, memo string, amount *big.Int, sender common.Address) (*transfertypes.MsgTransfer, error) {
	// Default to 100 blocks timeout
	timeoutHeight := types.NewHeight(0, 100)

	// Use instance to prevent errors on denom or amount
	token := sdk.Coin{
		Denom:  denom,
		Amount: math.NewIntFromBigInt(amount),
	}

	// Validate the token before creating the message
	if err := token.Validate(); err != nil {
		return nil, err
	}

	msg := &transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    OsmosisChannelId,
		Token:            token,
		Sender:           sdk.AccAddress(sender.Bytes()).String(), // convert to bech32 format
		Receiver:         OsmosisXCSContract,                      // The XCS contract address on Osmosis
		TimeoutHeight:    timeoutHeight,
		TimeoutTimestamp: transfertypes.DefaultRelativePacketTimeoutTimestamp,
		Memo:             memo,
	}

	// Validate the message before returning
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
