// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
//
// Osmosis package contains the logic of the Osmosis outpost on the Evmos chain.
// This outpost uses the ics20 precompile to relay IBC packets to the Osmosis
// chain, targeting the Cross-Chain Swap Contract V1 (XCS V1)

package osmosis

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/evmos/evmos/v16/precompiles/ics20"
)

const (
	// SwapMethod is the name of the swap method.
	SwapMethod = "swap"
)

const (
	// NextMemo is the memo to use after the swap of the token in the IBC packet
	// built on the Osmosis chain.
	NextMemo = ""
)

// Swap is a transaction that swap tokens on the Osmosis chain.
func (p Precompile) Swap(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	contract *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	swapPacketData, err := ParseSwapPacketData(method, args)
	if err != nil {
		return nil, err
	}

	input := swapPacketData.Input
	output := swapPacketData.Output
	amount := swapPacketData.Amount
	swapReceiver := swapPacketData.SwapReceiver

	// The provided sender address should always be equal to the origin address.
	// In case the contract caller address is the same as the sender address provided,
	// update the sender address to be equal to the origin address.
	// Otherwise, if the provided sender address is different from the origin address,
	// return an error because is a forbidden operation
	sender, err := ics20.CheckOriginAndSender(contract, origin, swapPacketData.Sender)
	if err != nil {
		return nil, err
	}

	bondDenom := p.stakingKeeper.GetParams(ctx).BondDenom
	var inputDenom, outputDenom string

	// Case 1. Input has to be either the address of Osmosis or WEVMOS
	switch input {
	case p.wevmosAddress:
		inputDenom = bondDenom
	default:
		inputDenom, err = p.erc20Keeper.GetTokenDenom(ctx, input)
		if err != nil {
			return nil, err
		}
	}

	// Case 2. Output has to be either the address of Osmosis or WEVMOS
	switch output {
	case p.wevmosAddress:
		outputDenom = bondDenom
	default:
		outputDenom, err = p.erc20Keeper.GetTokenDenom(ctx, output)
		if err != nil {
			return nil, err
		}
	}

	evmosChannel := NewIBCChannel(transfertypes.PortID, swapPacketData.ChannelID)
	err = ValidateInputOutput(inputDenom, outputDenom, bondDenom, evmosChannel)
	if err != nil {
		return nil, err
	}

	// Retrieve Osmosis channel and port associated with Evmos transfer app. We need these information
	// to reconstruct the output denom in the Osmosis chain.

	channel, found := p.channelKeeper.GetChannel(ctx, evmosChannel.PortID, evmosChannel.ChannelID)
	if !found {
		return nil, errorsmod.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", evmosChannel.PortID, evmosChannel.ChannelID)
	}
	osmosisChannel := NewIBCChannel(
		channel.GetCounterparty().GetPortID(),
		channel.GetCounterparty().GetChannelID(),
	)

	outputOnOsmosis, err := ConvertToOsmosisRepresentation(outputDenom, bondDenom, evmosChannel, osmosisChannel)
	if err != nil {
		return nil, err
	}

	// We have to compute the receiver address on the Osmosis chain to have a recovery address.
	onFailedDelivery := CreateOnFailedDeliveryField(sdk.AccAddress(sender.Bytes()).String())
	packet := CreatePacketWithMemo(
		outputOnOsmosis,
		swapPacketData.SwapReceiver,
		swapPacketData.XcsContract,
		swapPacketData.SlippagePercentage,
		swapPacketData.WindowSeconds,
		onFailedDelivery,
		NextMemo,
	)

	err = packet.Validate()
	if err != nil {
		return nil, err
	}
	packetString := packet.String()

	timeoutTimestamp := ctx.BlockTime().Add(ics20.DefaultTimeoutMinutes * time.Minute).UnixNano()
	coin := sdk.Coin{Denom: inputDenom, Amount: math.NewIntFromBigInt(amount)}
	msg, err := ics20.CreateAndValidateMsgTransfer(
		evmosChannel.PortID,
		evmosChannel.ChannelID,
		coin,
		sdk.AccAddress(sender.Bytes()).String(),
		swapPacketData.XcsContract,
		ics20.DefaultTimeoutHeight,
		uint64(timeoutTimestamp),
		packetString,
	)
	if err != nil {
		return nil, err
	}

	// No need to have authorization when the contract caller is the same as
	// origin (owner of funds) and the sender is the origin.
	accept, expiration, err := ics20.CheckAndAcceptAuthorizationIfNeeded(
		ctx,
		contract,
		origin,
		p.AuthzKeeper,
		msg,
	)
	if err != nil {
		return nil, err
	}

	// Execute the ICS20 Transfer.
	_, err = p.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	// Update grant only if is needed.
	if err := ics20.UpdateGrantIfNeeded(ctx, contract, p.AuthzKeeper, origin, expiration, accept); err != nil {
		return nil, err
	}

	// Emit the IBC transfer Event.
	if err := ics20.EmitIBCTransferEvent(
		ctx,
		stateDB,
		p.Events[ics20.EventTypeIBCTransfer],
		p.Address(),
		sender,
		msg.Receiver,
		msg.SourcePort,
		msg.SourceChannel,
		coin,
		packetString,
	); err != nil {
		return nil, err
	}

	// Emit the custom Swap Event.
	if err := p.EmitSwapEvent(ctx, stateDB, sender, input, output, amount, swapReceiver); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}
