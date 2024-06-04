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
	fmt.Println("IBCSendPacketCallback", sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, packetData, contractAddress, packetSenderAddress)
	contractHex := common.HexToAddress(contractAddress)
	// Checks if the contract supports ERC-165
	if err := k.DetectInterface(cachedCtx, OnSendPacketInterfaceID, contractHex); err != nil {
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

	_, err = k.evmKeeper.CallEVMWithData(cachedCtx, common.BytesToAddress(hexAddr), &contractHex, data, true)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) IBCOnAcknowledgementPacketCallback(cachedCtx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress, contractAddress, packetSenderAddress string) error {
	fmt.Println("IBCOnAcknowledgementPacketCallback", packet, acknowledgement, relayer, contractAddress, packetSenderAddress)
	contractHex := common.HexToAddress(contractAddress)
	// Checks if the contract supports ERC-165
	if err := k.DetectInterface(cachedCtx, OnAckPacketInterfaceID, contractHex); err != nil {
		fmt.Println("ack interface error", err)
		return err
	}

	ics20Packet, err := k.DecodeTransferPacketData(packet.Data)
	if err != nil {
		fmt.Println("ack decode error", err)
		return err
	}

	customICS20Packet := ICS20Packet{
		SourcePort:         packet.SourcePort,
		SourceChannel:      packet.SourceChannel,
		DestinationPort:    packet.DestinationPort,
		DestinationChannel: packet.DestinationChannel,
		TimeoutHeight:      packet.TimeoutHeight,
		TimeoutTimestamp:   packet.TimeoutTimestamp,
		Data:               ics20Packet,
	}

	data, err := k.ABI.Pack(IBCAcknowledgementPacketMethod, customICS20Packet, acknowledgement, common.BytesToAddress(relayer))
	if err != nil {
		fmt.Println("ack pack error", err)
		return err
	}

	prefix := strings.SplitN(packetSenderAddress, "1", 2)[0]
	hexAddr, err := sdk.GetFromBech32(packetSenderAddress, prefix)
	fmt.Println("ack hexAddr err", err)

	_, err = k.evmKeeper.CallEVMWithData(cachedCtx, common.BytesToAddress(hexAddr), &contractHex, data, true)
	if err != nil {
		fmt.Println("ack call error", err)
		return err
	}

	return nil

}

func (k Keeper) IBCOnTimeoutPacketCallback(cachedCtx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress, contractAddress, packetSenderAddress string) error {
	fmt.Println("IBCOnTimeoutPacketCallback", packet, relayer, contractAddress, packetSenderAddress)
	contractHex := common.HexToAddress(contractAddress)
	// Checks if the contract supports ERC-165
	if err := k.DetectInterface(cachedCtx, OnTimeoutPacketInterfaceID, contractHex); err != nil {
		return err
	}

	ics20Packet, err := k.DecodeTransferPacketData(packet.Data)
	if err != nil {
		return err
	}

	customICS20Packet := ICS20Packet{
		SourcePort:         packet.SourcePort,
		SourceChannel:      packet.SourceChannel,
		DestinationPort:    packet.DestinationPort,
		DestinationChannel: packet.DestinationChannel,
		TimeoutHeight:      packet.TimeoutHeight,
		TimeoutTimestamp:   packet.TimeoutTimestamp,
		Data:               ics20Packet,
	}

	data, err := k.ABI.Pack(IBCTimeoutPacketMethod, customICS20Packet, common.HexToAddress(packetSenderAddress))
	if err != nil {
		return err
	}

	prefix := strings.SplitN(packetSenderAddress, "1", 2)[0]
	hexAddr, err := sdk.GetFromBech32(packetSenderAddress, prefix)

	_, err = k.evmKeeper.CallEVMWithData(cachedCtx, common.BytesToAddress(hexAddr), &contractHex, data, true)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) IBCReceivePacketCallback(cachedCtx sdk.Context, packet exported.PacketI, ack exported.Acknowledgement, contractAddress string) error {
	fmt.Println("IBCReceivePacket", packet, contractAddress, ack)
	contractHex := common.HexToAddress(contractAddress)
	// Checks if the contract supports ERC-165
	if err := k.DetectInterface(cachedCtx, OnRecvPacketInterfaceID, contractHex); err != nil {
		fmt.Println("receive interface error", err)
		return err
	}

	channelPacket := packet.(channeltypes.Packet)
	ics20Packet, err := k.DecodeTransferPacketData(channelPacket.Data)
	if err != nil {
		fmt.Println("receive decode error", err)
		return err
	}

	customICS20Packet := ICS20Packet{
		SourcePort:         channelPacket.SourcePort,
		SourceChannel:      channelPacket.SourceChannel,
		DestinationPort:    channelPacket.DestinationPort,
		DestinationChannel: channelPacket.DestinationChannel,
		TimeoutHeight:      channelPacket.TimeoutHeight,
		TimeoutTimestamp:   channelPacket.TimeoutTimestamp,
		Data:               ics20Packet,
	}

	data, err := k.ABI.Pack(IBCReceivePacketMethod, customICS20Packet, common.Address{})
	if err != nil {
		fmt.Println("receive pack error", err)
		return err
	}

	_, err = k.evmKeeper.CallEVMWithData(cachedCtx, common.Address{}, &contractHex, data, true)
	if err != nil {
		fmt.Println("receive call error", err)
		return err
	}

	return nil
}
