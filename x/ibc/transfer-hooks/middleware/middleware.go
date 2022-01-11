package middleware

import (
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channelkeeper "github.com/cosmos/ibc-go/v3/modules/core/04-channel/keeper"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/tharsis/evmos/x/ibc/transfer-hooks/types"
)

var _ transfertypes.ICS4Wrapper = &Middleware{}

// Middleware defines the IBC fungible transfer hooks middleware
type Middleware struct {
	ChannelKeeper channelkeeper.Keeper
	hooks         types.TransferHooks
}

// NewMiddleware creates a new IBC transfer hooks Middleware instance
func NewMiddleware() Middleware {
	return Middleware{}
}

// Logger returns a module-specific logger.
func (Middleware) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+host.ModuleName+"-"+types.ModuleName)
}

// SetHooks sets
func (m *Middleware) SetHooks(th types.TransferHooks) *Middleware {
	if m.hooks != nil {
		panic("cannot set hooks twice")
	}

	m.hooks = th
	return m
}

// IsTransferHooksEnabled returns whether transfer hooks logic should be run.
func (m Middleware) IsTransferHooksEnabled() bool {
	return m.hooks != nil
}

// SendPacket implements the ICS4Wrapper interface from the transfer module.
// It calls the channel keeper SendPacket function directly.
func (m Middleware) SendPacket(ctx sdk.Context, channelCap *capabilitytypes.Capability, packet exported.PacketI) error {
	return m.ChannelKeeper.SendPacket(ctx, channelCap, packet)
}
