package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v6/modules/core/exported"
	"github.com/ethereum/go-ethereum/common"
)

// SendPacket defines a wrapper function for the channel Keeper's SendPacket
// function in order to expose it to the dIBC EVM Extension.
func (k Keeper) SendPacket(
	ctx sdk.Context,
	// TODO: check sender or use ADR-08
	packet ibcexported.PacketI,
) error {
	portID := packet.GetSourcePort()
	channelID := packet.GetSourceChannel()
	capName := host.ChannelCapabilityPath(portID, channelID)

	chanCap, ok := k.GetCapability(ctx, capName)
	if !ok {
		return sdkerrors.Wrapf(
			channeltypes.ErrChannelCapabilityNotFound,
			"could not retrieve channel capability at: %s", capName,
		)
	}

	// TODO: is ICS4Wrapper necessary here?
	return k.channelKeeper.SendPacket(ctx, chanCap, packet)
}

// OnRecvPacket calls the OnRecvPacket
func (k Keeper) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) channeltypes.Acknowledgement {
	// verify the destination port and channel
	if err := k.ValidatedIBCChannelParams(
		ctx, packet.GetDestPort(), packet.GetDestChannel(),
	); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	//  execute onRecv ABI call from contract that implement the IBC callback
	// FIXME: use constant instead

	// TODO: for this to work the packet needs to be 1:1 with the solidity struct
	data, err := k.callbacks.Methods["onRecvPacket"].Inputs.Pack(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// TODO: does the relayer pay for the fee?
	// do we handle permissions in the case that a method can only be executed by another contract?
	// consider creating a module account that would pay for the fee and is
	// the only one allowed to execute the tx

	sender := common.Address{}
	contract := common.HexToAddress(packet.GetDestPort())

	// TODO: check that the contract implements the IBCCallback ABI. This could be
	// done using the Interface method

	// TODO: handle case where the contract doesn't implement the ABI
	res, err := k.evmKeeper.CallEVMWithData(ctx, sender, &contract, data, true)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// revert the EVM tx and return the error so that the user can retry
	if res.VmError != "" {
		// TODO: error needs to be ABCI error
		return channeltypes.NewErrorAcknowledgement(errors.New(res.VmError))
	}

	// TODO: should the result contain the receipt or something else besides the data?
	return channeltypes.NewResultAcknowledgement(res.Ret)
}

// WriteAcknowledgement defines a wrapper function for the channel Keeper's function
// in order to expose it to the dIBC IBC handler.
func (k Keeper) WriteAcknowledgement(
	ctx sdk.Context,
	packet ibcexported.PacketI,
	acknowledgement []byte,
) error {
	portID := packet.GetDestPort()
	channelID := packet.GetDestChannel()
	capName := host.ChannelCapabilityPath(portID, channelID)
	chanCap, ok := k.GetCapability(ctx, capName)
	if !ok {
		return sdkerrors.Wrapf(channeltypes.ErrChannelCapabilityNotFound, "could not retrieve channel capability at: %s", capName)
	}

	// NOTE: a ResultAcknowledgement is alway be successful
	ack := channeltypes.NewResultAcknowledgement(acknowledgement)

	return k.channelKeeper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}
