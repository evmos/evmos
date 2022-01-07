package transferhooks

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/tharsis/evmos/x/ibc/transfer-hooks/keeper"
	"github.com/tharsis/evmos/x/ibc/transfer-hooks/types"
)

var _ porttypes.IBCModule = &IBCModule{}

// IBCModule implements the ICS26 callbacks for the fee middleware given the fee keeper and the underlying application.
type IBCModule struct {
	keeper keeper.Keeper
	app    porttypes.IBCModule
}

// NewIBCModule creates a new IBCModule given the keeper and underlying application
func NewIBCModule(k keeper.Keeper, app porttypes.IBCModule) IBCModule {
	return IBCModule{
		keeper: k,
		app:    app,
	}
}

// OnChanOpenInit implements the IBCModule interface
func (im IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) error {
	mwVersion, appVersion := channeltypes.SplitChannelVersion(version)
	// Since it is valid for fee version to not be specified, the above middleware version may be for a middleware
	// lower down in the stack. Thus, if it is not a fee version we pass the entire version string onto the underlying
	// application.
	// If an invalid fee version was passed, we expect the underlying application to fail on its version negotiation.
	if mwVersion == types.Version {
		im.keeper.SetTransferHooksEnabled(ctx, portID, channelID)
	} else {
		// middleware version is not the expected version for this midddleware. Pass the full version string along,
		// if it not valid version for any other lower middleware, an error will be returned by base application.
		appVersion = version
	}

	// call underlying app's OnChanOpenInit callback with the appVersion
	return im.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID,
		chanCap, counterparty, appVersion)
}

// OnChanOpenTry implements the IBCModule interface
func (im IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (version string, err error) {
	mwVersion, _ := channeltypes.SplitChannelVersion(version)
	cpMwVersion, cpAppVersion := channeltypes.SplitChannelVersion(counterpartyVersion)

	// Since it is valid for fee version to not be specified, the above middleware version may be for a middleware
	// lower down in the stack. Thus, if it is not a fee version we pass the entire version string onto the underlying
	// application.
	// If an invalid fee version was passed, we expect the underlying application to fail on its version negotiation.
	if mwVersion == types.Version || cpMwVersion == types.Version {
		if cpMwVersion != mwVersion {
			return "", sdkerrors.Wrapf(types.ErrInvalidVersion, "fee versions do not match. self version: %s, counterparty version: %s", mwVersion, cpMwVersion)
		}

		im.keeper.SetTransferHooksEnabled(ctx, portID, channelID)
	} else {
		// middleware versions are not the expected version for this middleware. Pass the full version strings along,
		// if it not valid version for any other lower middleware, an error will be returned by base application.
		cpAppVersion = counterpartyVersion
	}

	// call underlying app's OnChanOpenTry callback with the app versions
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID,
		chanCap, counterparty, cpAppVersion)
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyVersion string,
) error {
	// If handshake was initialized with fee enabled it must complete with fee enabled.
	// If handshake was initialized with fee disabled it must complete with fee disabled.
	cpAppVersion := counterpartyVersion
	if im.keeper.IsTransferHooksEnabled(ctx, portID, channelID) {
		var cpFeeVersion string
		cpFeeVersion, cpAppVersion = channeltypes.SplitChannelVersion(counterpartyVersion)

		if cpFeeVersion != types.Version {
			return sdkerrors.Wrapf(types.ErrInvalidVersion, "expected counterparty version: %s, got: %s", types.Version, cpFeeVersion)
		}
	}
	// call underlying app's OnChanOpenAck callback with the counterparty app version.
	return im.app.OnChanOpenAck(ctx, portID, channelID, cpAppVersion)
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// call underlying app's OnChanOpenConfirm callback.
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// delete fee enabled on channel
	// and refund any remaining fees escrowed on channel
	im.keeper.DeleteTransferHooksEnabled(ctx, portID, channelID)
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// delete fee enabled on channel
	// and refund any remaining fees escrowed on channel
	im.keeper.DeleteTransferHooksEnabled(ctx, portID, channelID)
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnRecvPacket implements the IBCModule interface.
// If fees are not enabled, this callback will default to the ibc-core packet callback
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	ack := im.app.OnRecvPacket(ctx, packet, relayer)
	if ack.Success() && im.keeper.IsTransferHooksEnabled(ctx, packet.DestinationPort, packet.DestinationChannel) {
		// TODO: add Transfer Hook Recv logic

		var data transfertypes.FungibleTokenPacketData
		if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
			err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
			return channeltypes.NewErrorAcknowledgement(err.Error())
		}

		// parse the transfer amount
		transferAmount, ok := sdk.NewIntFromString(data.Amount)
		if !ok {
			err := sdkerrors.Wrapf(transfertypes.ErrInvalidAmount, "unable to parse transfer amount (%s) into sdk.Int", data.Amount)
			return channeltypes.NewErrorAcknowledgement(err.Error())
		}

		token := sdk.NewCoin(data.Denom, transferAmount)
		isSource := transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), data.Denom)

		// unmarshal packet
		im.keeper.AfterRecvTransfer(ctx, packet.DestinationPort, packet.DestinationChannel, token, data.Receiver, isSource)
	}

	return ack
}

// OnAcknowledgementPacket implements the IBCModule interface
// If fees are not enabled, this callback will default to the ibc-core packet callback
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	if err := im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer); err != nil {
		return err
	}

	var ack channeltypes.Acknowledgement
	if err := transfertypes.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}

	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	// parse the transfer amount
	transferAmount, ok := sdk.NewIntFromString(data.Amount)
	if !ok {
		return sdkerrors.Wrapf(transfertypes.ErrInvalidAmount, "unable to parse transfer amount (%s) into sdk.Int", data.Amount)
	}

	token := sdk.NewCoin(data.Denom, transferAmount)

	sender, err := sdk.AccAddressFromBech32(data.Sender)
	if err != nil {
		return err
	}

	fullDenomPath := data.Denom
	// deconstruct the token denomination into the denomination trace info
	// to determine if the sender is the source chain
	if strings.HasPrefix(token.Denom, "ibc/") {
		fullDenomPath, err = im.keeper.DenomPathFromHash(ctx, token.Denom)
		if err != nil {
			return err
		}
	}

	isSource := transfertypes.SenderChainIsSource(packet.SourcePort, packet.SourceChannel, fullDenomPath)

	im.keeper.AfterSendTransferAcked(ctx, packet.SourcePort, packet.SourceChannel, token, sender, data.Receiver, isSource)
	return nil
}

// OnTimeoutPacket implements the IBCModule interface
// If fees are not enabled, this callback will default to the ibc-core packet callback
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}
