package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/tharsis/evmos/v5/x/recovery/types"
)

var _ transfertypes.ICS4Wrapper = Keeper{}

// Keeper struct
type Keeper struct {
	paramstore     paramtypes.Subspace
	accountKeeper  types.AccountKeeper
	bankKeeper     types.BankKeeper
	ics4Wrapper    porttypes.ICS4Wrapper
	channelKeeper  types.ChannelKeeper
	transferKeeper types.TransferKeeper
	claimsKeeper   types.ClaimsKeeper
}

// NewKeeper returns keeper
func NewKeeper(
	ps paramtypes.Subspace,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	ck types.ChannelKeeper,
	tk types.TransferKeeper,
	claimsKeeper types.ClaimsKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		paramstore:     ps,
		accountKeeper:  ak,
		bankKeeper:     bk,
		channelKeeper:  ck,
		transferKeeper: tk,
		claimsKeeper:   claimsKeeper,
	}
}

// SetICS4Wrapper sets the ICS4 wrapper to the keeper.
// It panics if already set
func (k *Keeper) SetICS4Wrapper(ics4Wrapper porttypes.ICS4Wrapper) {
	if k.ics4Wrapper != nil {
		panic("ICS4 wrapper already set")
	}

	k.ics4Wrapper = ics4Wrapper
}

// Logger returns logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// IBC callbacks and transfer handlers

// SendPacket implements the ICS4Wrapper interface from the transfer module.
// It calls the underlying SendPacket function directly to move down the middleware stack.
func (k Keeper) SendPacket(ctx sdk.Context, channelCap *capabilitytypes.Capability, packet exported.PacketI) error {
	return k.ics4Wrapper.SendPacket(ctx, channelCap, packet)
}

// WriteAcknowledgement implements the ICS4Wrapper interface from the transfer module.
// It calls the underlying WriteAcknowledgement function directly to move down the middleware stack.
func (k Keeper) WriteAcknowledgement(ctx sdk.Context, channelCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, channelCap, packet, ack)
}
