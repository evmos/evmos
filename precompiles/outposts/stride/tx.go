package stride

import (
	"embed"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"log"
)

// Embed memo json file to the executable binary. Needed when importing as dependency.
//
//go:embed memo.json
var memoF embed.FS

const (
	// StrideChannelID is the channel ID for the Stride channel
	StrideChannelID = "channel-25"
	// LiquidStakeEvmosMethod is the method name of the LiquidStakeEvmos method
	LiquidStakeEvmosMethod = "liquidStakeEvmos"
)

// LiquidStakeEvmos is a transaction that liquid stakes Evmos using
// a ICS20 transfer with a custom memo field that will trigger Stride's Autopilot middleware
func (p Precompile) LiquidStakeEvmos(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	coin, receiverAddress, err := CreateLiquidStakeEvmosPacket(args, p.stakingKeeper.BondDenom(ctx))
	if err != nil {
		return nil, err
	}

	// Check if the channel with Stride is found
	_, found := p.channelKeeper.GetChannel(ctx, transfertypes.PortID, StrideChannelID)
	if !found {
		return nil, channeltypes.ErrChannelNotFound
	}

	memo := p.createLiquidStakeMemo(receiverAddress)

	// Build the MsgTransfer with the memo and coin
	msg, err := NewMsgTransfer(StrideChannelID, sdk.AccAddress(origin.Bytes()).String(), receiverAddress, memo, coin)
	if err != nil {
		return nil, err
	}

	// Execute the ICS20 Transfer
	_, err = p.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	// Emit the IBC transfer Event
	// TODO: Figure out if we want a more custom event here to signal Autopilot usage
	if err = p.EmitIBCTransferEvent(
		ctx,
		stateDB,
		origin,
		msg.Receiver,
		msg.SourcePort,
		msg.SourceChannel,
		msg.Token,
		msg.Memo,
	); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// createLiquidStakeMemo creates the memo for the LiquidStakeEvmos packet
func (p Precompile) createLiquidStakeMemo(receiverAddress string) string {
	// Read the JSON memo from the file
	data, err := memoF.ReadFile("memo.json")
	if err != nil {
		log.Fatalf("Failed to read JSON memo: %v", err)
	}

	// Replace the placeholder with the receiver address
	return fmt.Sprintf(string(data), receiverAddress)
}
