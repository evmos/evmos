package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
)

// CallbackPacketData Implemented by any packet data type that wants to support PacketActor callbacks
// PacketActor's will be unable to act on any packet data type that does not implement
// this interface.
type CallbackPacketData interface {
	// GetSourceCallbackAddress should return the callback address of a packet data on the source chain.
	// This may or may not be the sender of the packet. If no source callback address exists for the packet,
	// an empty string may be returned.
	GetSourceCallbackAddress() string

	// GetDestCallbackAddress should return the callback address of a packet data on the destination chain.
	// This may or may not be the receiver of the packet. If no dest callback address exists for the packet,
	// an empty string may be returned.
	GetDestCallbackAddress() string

	// UserDefinedGasLimit allows the sender of the packet to define inside the packet data
	// a gas limit for how much the ADR-8 callbacks can consume. If defined, this will be passed
	// in as the gas limit so that the callback is guaranteed to complete within a specific limit.
	// On recvPacket, a gas-overflow will just fail the transaction allowing it to timeout on the sender side.
	// On ackPacket and timeoutPacket, a gas-overflow will reject state changes made during callback but still
	// commit the transaction. This ensures the packet lifecycle can always complete.
	// If the packet data returns 0, the remaining gas limit will be passed in (modulo any chain-defined limit)
	// Otherwise, we will set the gas limit passed into the callback to the `min(ctx.GasLimit, UserDefinedGasLimit())`
	UserDefinedGasLimit() uint64
}

type IBCActor interface {
	// OnChannelOpen will be called on the IBCActor when the channel opens
	// this will happen either on ChanOpenAck or ChanOpenConfirm
	OnChannelOpen(ctx sdk.Context, portID, channelID, version string)

	// OnChannelClose will be called on the IBCActor if the channel closes
	// this will be called on either ChanCloseInit or ChanCloseConfirm and if the channel handshake fails on our end
	// NOTE: currently the channel does not automatically close if the counterparty fails the handhshake so actors must be prepared for an OpenInit to never return a callback for the time being
	OnChannelClose(ctx sdk.Context, portID, channelID string)

	// PacketActor - IBCActor must also implement PacketActor interface
	PacketActor
}

// PacketActor is split out into its own separate interface since implementors may choose
// to only support callbacks for packet methods rather than supporting the full IBCActor interface
type PacketActor interface {
	// OnRecvPacket will be called on the IBCActor after the IBC Application
	// handles the RecvPacket callback if the packet has an IBC Actor as a receiver.
	OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error

	// OnAcknowledgmentPacket will be called on the IBC Actor
	// after the IBC Application handles its own OnAcknowledgementPacket callback
	OnAcknowledgmentPacket(
		ctx sdk.Context,
		packet channeltypes.Packet,
		ack exported.Acknowledgement,
		relayer sdk.AccAddress,
	) error

	// OnTimeoutPacket will be called on the IBC Actor
	// after the IBC Application handles its own OnTimeoutPacket callback
	OnTimeoutPacket(
		ctx sdk.Context,
		packet channeltypes.Packet,
		relayer sdk.AccAddress,
	) error
}
