package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/tharsis/evmos/x/withdraw/types"
)

var _ transfertypes.ICS4Wrapper = Keeper{}

// Keeper struct
type Keeper struct {
	cdc            codec.Codec
	accountKeeper  types.AccountKeeper
	bankKeeper     types.BankKeeper
	transferKeeper types.TransferKeeper
}

// NewKeeper returns keeper
func NewKeeper(
	cdc codec.Codec,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	tk types.TransferKeeper,
) *Keeper {

	return &Keeper{
		cdc:            cdc,
		accountKeeper:  ak,
		bankKeeper:     bk,
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
	return k.transferKeeper.SendPacket(ctx, channelCap, packet)
}
