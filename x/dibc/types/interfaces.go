package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capability "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"

	connection "github.com/cosmos/ibc-go/v6/modules/core/03-connection/types"
	channel "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v6/modules/core/exported"

	"github.com/evmos/evmos/v13/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v13/x/evm/types"
)

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channel.Channel, found bool)
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
	SendPacket(ctx sdk.Context, channelCap *capability.Capability, packet ibcexported.PacketI) error
	WriteAcknowledgement(ctx sdk.Context, channelCap *capability.Capability, packet ibcexported.PacketI, acknowledgement ibcexported.Acknowledgement) error
	ChanOpenInit(
		ctx sdk.Context,
		order channel.Order,
		connectionHops []string,
		portID string,
		portCap *capability.Capability,
		counterparty channel.Counterparty,
		version string,
	) (string, *capability.Capability, error)
	ChanCloseInit(ctx sdk.Context, portID, channelID string, chanCap *capability.Capability) error
	TimeoutExecuted(ctx sdk.Context, channelCap *capability.Capability, packet ibcexported.PacketI) error
	WriteOpenInitChannel(ctx sdk.Context, portID, channelID string, order channel.Order, connectionHops []string, counterparty channel.Counterparty, version string)
}

// ClientKeeper defines the expected IBC client keeper
type ClientKeeper interface {
	GetClientConsensusState(ctx sdk.Context, clientID string) (connection ibcexported.ConsensusState, found bool)
}

// ConnectionKeeper defines the expected IBC connection keeper
type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (connection connection.ConnectionEnd, found bool)
}

// PortKeeper defines the expected IBC port keeper
type PortKeeper interface {
	BindPort(ctx sdk.Context, portID string) *capability.Capability
}

// EVMKeeper defines the expected EVM keeper interface used on erc20
type EVMKeeper interface {
	GetParams(ctx sdk.Context) evmtypes.Params
	GetAccountWithoutBalance(ctx sdk.Context, addr common.Address) *statedb.Account
	EstimateGas(c context.Context, req *evmtypes.EthCallRequest) (*evmtypes.EstimateGasResponse, error)
	ApplyMessage(ctx sdk.Context, msg core.Message, tracer vm.EVMLogger, commit bool) (*evmtypes.MsgEthereumTxResponse, error)
}
