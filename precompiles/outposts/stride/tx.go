package stride

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tendermint "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

const (
	// LiquidStakeEvmosMethod is the method name of the LiquidStakeEvmos method
	LiquidStakeEvmosMethod = "liquidStakeEvmos"
)

func (p Precompile) LiquidStakeEvmos(
	ctx sdk.Context,
	origin common.Address, // the EOA that signs
	contract *vm.Contract, // contract address right before the precompile address
	stateDB vm.StateDB, // to emit events
	method *abi.Method, // the method name and args
	args []interface{}, // the arguments from the function (function parameters)
) ([]byte, error) {
	coin, receiverAddress, err := CreateLiquidStakeEvmosPacket(args, p.stakingKeeper.BondDenom(ctx))
	if err != nil {
		return nil, err
	}

	memo := p.createLiquidStakeMemo(receiverAddress)

	// TODO: Some channel discovery logic here to find the correct channel with Stride
	channels := p.channelKeeper.GetAllChannels(ctx)
	for _, channel := range channels {
		if channel.State == 3 && channel.PortId == "transfer" {
			_, clientState, err := p.channelKeeper.GetChannelClientState(ctx, channel.PortId, channel.ChannelId)
			if err != nil {
				return nil, err
			}
			tendermintClientState := clientState.(*tendermint.ClientState)
			// TODO: Add the chain-id for stride here
			if tendermintClientState.ChainId == "" {

			}
		}
	}

	bech32Origin := sdk.AccAddress(origin.Bytes()).String()

	// Build the MsgTransfer with the memo and coin
	msg, err := NewMsgTransfer("", bech32Origin, receiverAddress, memo, coin)
	if err != nil {
		return nil, err
	}

	_, err = p.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	if err = p.EmitIBCTransferEvent(
		ctx,
		stateDB,
		s,
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
// TODO: there a better here to do this string building
func (p Precompile) createLiquidStakeMemo(receiverAddress string) string {
	template := `{
			"autopilot":{
				"receiver": "%s",
				"stakeibc":{
					"action": "LiquidStake",
				}
			}
		}`
	return fmt.Sprintf(template, receiverAddress)
}
