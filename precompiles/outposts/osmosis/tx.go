package osmosis

import (
	"embed"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

// Embed memo json file to the executable binary. Needed when importing as dependency.
//
//go:embed memo.json
var memoF embed.FS

const (
	// OsmosisChannelId defines the channel id for the Osmosis IBC channel
	OsmosisChannelId = "channel-0"
	// OsmosisXCSContract defines the contract address for the Osmosis XCS contract
	OsmosisXCSContract = "osmo1xcsjj7g9qf6qy8w4xg2j3q4q3k6x5q2x9k5x2e"
)

const (
	// SwapMethod defines the ABI method name for the Osmosis Swap function
	SwapMethod = "swap"
)

// Swap swaps the given base denom for the given target denom on Osmosis and returns
// the newly swapped tokens to the receiver address.
func (p Precompile) Swap(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	amount, inputDenom, outputDenom, receiverAddress, err := CreateSwapPacketData(args, ctx, p.bankKeeper, p.erc20Keeper)
	if err != nil {
		return nil, err
	}

	// TODO: Include case where there is a SC calling into the Precompile

	// Create the memo field for the Swap from the JSON file
	memo, err := createSwapMemo(outputDenom, receiverAddress)
	if err != nil {
		return nil, err
	}

	// Create the IBC Transfer message
	msg, err := NewMsgTransfer(inputDenom, memo, amount, origin)
	if err != nil {
		return nil, err
	}

	// Send the IBC Transfer message
	_, err = p.transferKeeper.Transfer(ctx, msg)
	if err != nil {
		return nil, err
	}

	// Emit the ICS20 Transfer Event
	if err := p.EmitIBCTransferEvent(ctx, stateDB, origin, amount, inputDenom, memo); err != nil {
		return nil, err
	}

	receiverAccAddr, err := sdk.AccAddressFromBech32(receiverAddress)
	if err != nil {
		return nil, err
	}

	// Emit the Osmosis Swap Event
	// TODO: Check if the chainPrefix extraction works
	if err := p.EmitSwapEvent(ctx, stateDB, origin, common.BytesToAddress(receiverAccAddr), amount, inputDenom, outputDenom, ""); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// createSwapMemo creates a memo for the swap transaction
func createSwapMemo(outputDenom, receiverAddress string) (string, error) {
	// Read the JSON memo from the file
	data, err := memoF.ReadFile("memo.json")
	if err != nil {
		return "", fmt.Errorf("failed to read JSON memo: %v", err)
	}

	return fmt.Sprintf(string(data), OsmosisXCSContract, outputDenom, receiverAddress), nil
}
