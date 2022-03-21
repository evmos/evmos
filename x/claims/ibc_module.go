package claims

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/tharsis/evmos/v2/ibc"
	"github.com/tharsis/evmos/v2/x/claims/keeper"
)

var _ porttypes.IBCModule = &IBCModule{}

// IBCModule implements the ICS26 callbacks for the transfer middleware given
// the claim keeper and the underlying application.
type IBCModule struct {
	*ibc.Module
	keeper keeper.Keeper
}

// NewIBCModule creates a new IBCModule given the keeper and underlying application
func NewIBCModule(k keeper.Keeper, app porttypes.IBCModule) IBCModule {
	return IBCModule{
		Module: ibc.NewModule(app),
		keeper: k,
	}
}

// OnRecvPacket implements the IBCModule interface.
// If fees are not enabled, this callback will default to the ibc-core packet callback
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	ack := im.Module.OnRecvPacket(ctx, packet, relayer)

	// return if the acknowledgement is an error ACK
	if !ack.Success() {
		return ack
	}

	return im.keeper.OnRecvPacket(ctx, packet, ack)
}

// OnAcknowledgementPacket implements the IBCModule interface
// If fees are not enabled, this callback will default to the ibc-core packet callback
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	if err := im.Module.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer); err != nil {
		return err
	}

	return im.keeper.OnAcknowledgementPacket(ctx, packet, acknowledgement)
}
