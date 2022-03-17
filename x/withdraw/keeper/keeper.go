package keeper

import (
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/tharsis/evmos/v2/x/withdraw/types"
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

// GetIBCDenomSource returns the self port and channel of the IBC denomination (i.e port and channel on Evmos for the voucher).
// It returns an error if:
// - the the denomination is invalid
// - the denom trace or source channel are not found on the store
// - self port or channel ID are invalid
func (k Keeper) GetIBCDenomSelfIdentifiers(ctx sdk.Context, denom, sender string) (counterpartyPort, counterpartyChannel string, err error) {
	ibcDenom := strings.SplitN(denom, "/", 2)
	if len(ibcDenom) < 2 {
		return "", "", sdkerrors.Wrap(transfertypes.ErrInvalidDenomForTransfer, denom)
	}

	hash, err := transfertypes.ParseHexHash(ibcDenom[1])
	if err != nil {
		return "", "", sdkerrors.Wrapf(
			err,
			"failed to withdraw IBC vouchers back to sender '%s' in the corresponding IBC chain", sender,
		)
	}

	denomTrace, found := k.transferKeeper.GetDenomTrace(ctx, hash)
	if !found {
		return "", "", sdkerrors.Wrapf(
			transfertypes.ErrTraceNotFound,
			"failed to withdraw IBC vouchers back to sender '%s' in the corresponding IBC chain", sender,
		)
	}

	path := strings.Split(denomTrace.Path, "/")
	if len(path)%2 != 0 {
		// safety check: shouldn't occur
		return "", "", sdkerrors.Wrapf(
			transfertypes.ErrInvalidDenomForTransfer,
			"invalid denom (%s) trace path %s", denomTrace.BaseDenom, denomTrace.Path,
		)
	}

	counterpartyPortID := path[0]
	counterpartyChannelID := path[1]

	channel, found := k.channelKeeper.GetChannel(ctx, counterpartyPortID, counterpartyChannelID)
	if !found {
		return "", "", sdkerrors.Wrapf(
			channeltypes.ErrChannelNotFound,
			"port ID %s, channel ID %s", counterpartyPortID, counterpartyChannelID,
		)
	}

	sourcePort := channel.Counterparty.PortId
	sourceChannel := channel.Counterparty.ChannelId

	// NOTE: optimistic handshakes could cause unforeseen issues.
	// Safety check: verify that the source port and channel are valid
	if err := host.PortIdentifierValidator(sourcePort); err != nil {
		// shouldn't occur
		return "", "", sdkerrors.Wrapf(
			host.ErrInvalidID,
			"invalid port ID '%s': %s", sourcePort, err.Error(),
		)
	}

	if err := host.ChannelIdentifierValidator(sourceChannel); err != nil {
		// shouldn't occur
		return "", "", sdkerrors.Wrapf(
			channeltypes.ErrInvalidChannelIdentifier,
			"channel ID '%s': %s", sourceChannel, err.Error(),
		)
	}

	return counterpartyPort, counterpartyChannel, nil
}
