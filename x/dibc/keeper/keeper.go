package keeper

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capability "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	host "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v6/modules/core/exported"

	"github.com/evmos/evmos/v13/x/dibc/types"
	evmkeeper "github.com/evmos/evmos/v13/x/evm/keeper"
)

// Keeper maintains the link to data storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	// Protobuf codec
	cdc codec.BinaryCodec

	storeKey storetypes.StoreKey

	channelKeeper types.ChannelKeeper
	portKeeper    types.PortKeeper
	scopedKeeper  capabilitykeeper.ScopedKeeper
	evmKeeper     evmkeeper.Keeper
	callbacks     *abi.ABI
}

// NewKeeper creates a new dIBC Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	scopedKeeper capabilitykeeper.ScopedKeeper,
	portKeeper types.PortKeeper,
	channelKeeper types.ChannelKeeper,
	evmKeeper evmkeeper.Keeper,
) Keeper {
	return Keeper{
		storeKey:      key,
		cdc:           cdc,
		evmKeeper:     evmKeeper,
		channelKeeper: channelKeeper,
		portKeeper:    portKeeper,
		scopedKeeper:  scopedKeeper,
	}
}

// ChanOpenInit defines a wrapper function for the channel Keeper's function
// in order to expose it to the dIBC handler.
func (k Keeper) ChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	counterpartyPortID,
	version string,
) error {
	portCap, isBound := k.GetCapability(ctx, host.PortPath(portID))
	if isBound {
		return sdkerrors.Wrapf(porttypes.ErrPortNotFound, "port not found or already bound: %s", portID)
	}

	// create the counterparty and sets the channel identifier to be empty.
	counterparty := channeltypes.NewCounterparty(counterpartyPortID, "")

	channelID, chanCap, err := k.channelKeeper.ChanOpenInit(ctx, order, connectionHops, portID, portCap, counterparty, version)
	if err != nil {
		return err
	}

	if err := k.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
		return err
	}

	// TODO: write channel (?). Why is this not used in the Agoric implementation?
	k.channelKeeper.WriteOpenInitChannel(ctx, portID, channelID, order, connectionHops, counterparty, version)
	ctx.Logger().Info("channel open init callback succeeded", "channel-id", channelID, "version", version)

	return nil
}

// ChanCloseInit defines a wrapper function for the channel Keeper's function
// in order to expose it to the dIBC handler.
func (k Keeper) ChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	capName := host.ChannelCapabilityPath(portID, channelID)
	chanCap, ok := k.GetCapability(ctx, capName)
	if !ok {
		return sdkerrors.Wrapf(
			channeltypes.ErrChannelCapabilityNotFound,
			"could not retrieve channel capability at: %s", capName,
		)
	}

	if err := k.channelKeeper.ChanCloseInit(ctx, portID, channelID, chanCap); err != nil {
		return err
	}

	// We need to emit a channel event to notify the relayer.
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, channeltypes.AttributeValueCategory),
		),
	)

	return nil
}

// GetNextSequenceSend defines a wrapper function for the channel Keeper's function
// in order to expose it to the dIBC handler.
func (k Keeper) GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool) {
	return k.channelKeeper.GetNextSequenceSend(ctx, portID, channelID)
}

// GetChannel defines a wrapper function for the channel Keeper's function
// in order to expose it to the dIBC handler.
func (k Keeper) GetChannel(ctx sdk.Context, portID, channelID string) (channeltypes.Channel, bool) {
	return k.channelKeeper.GetChannel(ctx, portID, channelID)
}

// TimeoutExecuted defines a wrapper function for the channel Keeper's function
// in order to expose it to the dIBC IBC handler.
func (k Keeper) OnTimeoutPacket(ctx sdk.Context, packet ibcexported.PacketI) error {
	// verify the destination port and channel
	if err := k.ValidatedIBCChannelParams(
		ctx, packet.GetDestPort(), packet.GetDestChannel(),
	); err != nil {
		return err
	}

	//  execute OnTimeout ABI call from contract that implement the IBC callback
	// FIXME: use constant instead
	data, err := k.callbacks.Methods["onTimeoutPacket"].Inputs.Pack(packet)
	if err != nil {
		return err
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
		return err
	}

	// revert the EVM tx and return the error so that the user can retry
	if res.VmError != "" {
		return errors.New(res.VmError)
	}

	return nil
}

// ClaimCapability allows the dIBC module to claim a capability that IBC module
// passes to it
func (k Keeper) ClaimCapability(ctx sdk.Context, cap *capability.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}

func (k Keeper) GetCapability(ctx sdk.Context, name string) (*capability.Capability, bool) {
	return k.scopedKeeper.GetCapability(ctx, name)
}
