package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/ethereum/go-ethereum/common"
	"strings"
)

const (
	SupportsInterfaceQuery         = "supportsInterface"
	IBCSendPacketMethod            = "onSendPacket"
	IBCAcknowledgementPacketMethod = "onAcknowledgementPacket"
	IBCTimeoutPacketMethod         = "onTimeoutPacket"
	IBCReceivePacketMethod         = "onReceivePacket"
)

type ICS20Packet struct {
	SourcePort         string             `abi:"sourcePort"`
	SourceChannel      string             `abi:"sourceChannel"`
	DestinationPort    string             `abi:"destinationPort"`
	DestinationChannel string             `abi:"destinationChannel"`
	Data               ICS20EVMPacketData `abi:"data"`
	TimeoutHeight      clienttypes.Height `abi:"timeoutHeight"`
	TimeoutTimestamp   uint64             `abi:"timeoutTimestamp"`
}

func (k Keeper) IBCSendPacketCallback(cachedCtx sdk.Context, sourcePort string, sourceChannel string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64, packetData []byte, contractAddress, packetSenderAddress string) error {
	contractHex := common.HexToAddress(contractAddress)
	// Checks if the contract supports ERC-165
	if err := k.DetectInterface(cachedCtx, OnSendPacketInterfaceID, packetSenderAddress, contractHex); err != nil {
		return err
	}

	channel, found := k.channelKeeper.GetChannel(cachedCtx, sourcePort, sourceChannel)
	if !found {
		return fmt.Errorf("channel not found")
	}

	ics20Packet, err := k.DecodeTransferPacketData(packetData)
	if err != nil {
		return err
	}

	packet := ICS20Packet{
		SourcePort:         sourcePort,
		SourceChannel:      sourceChannel,
		DestinationPort:    channel.Counterparty.PortId,
		DestinationChannel: channel.Counterparty.ChannelId,
		TimeoutHeight:      timeoutHeight,
		TimeoutTimestamp:   timeoutTimestamp,
		Data:               ics20Packet,
	}

	data, err := k.ABI.Pack(IBCSendPacketMethod, packet, common.HexToAddress(packetSenderAddress))
	if err != nil {
		return err
	}

	prefix := strings.SplitN(packetSenderAddress, "1", 2)[0]
	hexAddr, err := sdk.GetFromBech32(packetSenderAddress, prefix)

	txResponse, err := k.CallEVMWithData(cachedCtx, common.BytesToAddress(hexAddr), &contractHex, data, true)
	if err != nil {
		fmt.Println("the error in call with evm", err)
		return err
	}

	fmt.Println(txResponse, "here response")
	return nil
}

func (k Keeper) IBCOnAcknowledgementPacketCallback(cachedCtx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress, contractAddress, packetSenderAddress string) error {
	fmt.Println("IBCOnAcknowledgementPacketCallback")

	return nil
}

func (k Keeper) IBCOnTimeoutPacketCallback(cachedCtx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress, contractAddress, packetSenderAddress string) error {
	fmt.Println("IBCOnTimeoutPacketCallback")
	return nil
}

func (k Keeper) IBCReceivePacketCallback(cachedCtx sdk.Context, packet exported.PacketI, ack exported.Acknowledgement, contractAddress string) error {
	fmt.Println("IBCReceivePacketCallback")
	return nil
}
