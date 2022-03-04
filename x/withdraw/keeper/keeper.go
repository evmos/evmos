package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/tharsis/evmos/x/withdraw/types"
)

var _ transfertypes.ICS4Wrapper = Keeper{}

// Keeper struct
type Keeper struct {
	paramstore     paramtypes.Subspace
	bankKeeper     types.BankKeeper
	ics4Wrapper    transfertypes.ICS4Wrapper
	transferKeeper types.TransferKeeper
}

// NewKeeper returns keeper
func NewKeeper(
	ps paramtypes.Subspace,
	bk types.BankKeeper,
	ics4Wrapper transfertypes.ICS4Wrapper,
	tk types.TransferKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		paramstore:     ps,
		bankKeeper:     bk,
		ics4Wrapper:    ics4Wrapper,
		transferKeeper: tk,
	}
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
